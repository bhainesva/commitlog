package commitlog

import (
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
)

// findPositionsToDelete is a helper function that returns a set of unused identifiers.
// It takes a map of asts by filename, a set of positions that contain active nodes
// and a map with type usage information
func findPositionsToDelete(astByName map[string]*ast.File, activePos map[token.Pos]struct{}, uses map[*ast.Ident]types.Object) map[token.Pos]struct{} {
	var (
		referencedTypePositions = map[token.Pos]struct{}{}
		deletionCandidates = map[token.Pos]struct{}{}
	)

	markParamsAsUsed := func(fields []*ast.Field) {
		for _, field := range fields {
			if fun, ok :=  field.Type.(*ast.FuncType); ok {
				for _, param := range fun.Params.List {
					for _, paramName := range param.Names {
						referencedTypePositions[paramName.Pos()] = struct{}{}
					}
				}
			}
		}
	}

	for _, file := range astByName {
		// Inspect the AST, mark identifiers for deletion that
		// 1. Have a nil entry in the typeinfo uses
		//                  AND
		// 2. Are not themselves referenced by any other Ident node through
		//    uses
		ast.Inspect(file, func(n ast.Node) bool {
			// Edge Cases
			// Too hard to check if implementers of an interface use all the params the method is specified to take
			// just leave them all in
			if n, ok := n.(*ast.InterfaceType); ok {
				markParamsAsUsed(n.Methods.List)
			}
			// Similarly, when a function takes a function argument, it would be difficult to
			// tell if the function passed actually uses the arguments that it takes. Leave them all in
			if n, ok := n.(*ast.FuncDecl); ok {
				markParamsAsUsed(n.Type.Params.List)
			}
			// We still strip unused params from function declarations in general
			// This is useful because it can allow you to remove the whole
			// type definition if it's otherwise unused.
			// However, if the function is intended to implement an interface, and some params are just unneeded
			// for the particular implementation, then it might lead to confusing / invalid code when they're stripped

			if n, ok := n.(*ast.Ident); ok {
				if n.Name == "main" {
					return true
				}

				if usedObj, ok := uses[n]; ok {
					if _, ok := activePos[n.Pos()]; ok {
						referencedTypePositions[usedObj.Pos()] = struct{}{}
						delete(deletionCandidates, usedObj.Pos())
					}
				} else {
					_, living := activePos[n.Pos()]
					if _, ok := referencedTypePositions[n.Pos()]; !ok || !living {
						deletionCandidates[n.Pos()] = struct{}{}
					}
				}
			}
			return true
		})
	}

	return deletionCandidates
}

// removeDeadCode takes a map of dst Files by filename and returns a similar map with a
// layer of dead code removed. It also returns a bool reporting whether any code was changed as
// a result
func removeDeadCode(trees map[string]*dst.File, fset *token.FileSet, decorators map[string]*decorator.Decorator) (map[string]*dst.File, bool, error) {
	var (
		codeDeleted = false
		astByName   = map[string]*ast.File{}
		astFiles    []*ast.File
		livingPOS = map[token.Pos]struct{}{}
	)

	for fn, d := range decorators {
		astFile := d.Ast.Nodes[trees[fn]].(*ast.File)
		astByName[fn] = astFile
		astFiles = append(astFiles, astFile)
	}

	// The DSTs may have been manipulated but the ASTs stored in the decorator map
	// are not updated to reflect that. This first pass figures out which AST nodes
	// correspond to dst nodes that still exist in the tree, and records their positions.
	for fn, tree := range trees {
		d := decorators[fn]
		dst.Inspect(tree, func(n dst.Node) bool {
			if n == nil {
				return true
			}

			ast := d.Ast.Nodes[n]

			livingPOS[ast.Pos()] = struct{}{}
			return true
		})
	}

	conf := types.Config{
		Importer:         importer.Default(),
		IgnoreFuncBodies: false,
		// Swallow errors, it's likely the input Files are invalid, for example
		// because of unused imports remaining when uncovered code using them has been removed
		Error:            func(error) {},
	}
	typesInfo := types.Info{
		Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object),
	}
	conf.Check("", fset, astFiles, &typesInfo)

	deletionCandidates := findPositionsToDelete(astByName, livingPOS, typesInfo.Uses)

	outFiles := map[string]*dst.File{}
	for name, file := range trees {
		// Sometimes we realize we want to delete a node while
		// visiting its children. These sets give information to the
		// post func to take care of that on the way back up the tree
		toDelete := map[dst.Node]struct{}{}
		toDeleteParentType := map[dst.Node]struct{}{}

		d := decorators[name]
		// Traverse the AST a second time, deleting any identifiers
		// marked for deletion in the first pass
		newTree := dstutil.Apply(file, func(c *dstutil.Cursor) bool {
			if _, ok := c.Node().(*dst.Ident); ok {
				astNode := d.Ast.Nodes[c.Node()]
				if _, ok := deletionCandidates[astNode.Pos()]; !ok {
					return true
				}

				parent := c.Parent()

				switch t := parent.(type) {
				case *dst.Field:
					// Only one field declared, delete whole row
					if len(t.Names) == 1 {
						toDelete[parent] = struct{}{}
						return false
					} else { // delete just this identifier
						codeDeleted = true
						c.Delete()
						return false
					}
				case *dst.TypeSpec:
					// In this case the node we want to delete isn't the direct
					// parent of the identifier. This struct is used to indicate
					// we should continue up the tree until we find a GenDecl
					toDeleteParentType[parent] = struct{}{}
					return false
				}
			}

			return true
		}, func(c *dstutil.Cursor) bool {
			node := c.Node()
			if _, ok := toDelete[node]; ok {
				delete(toDelete, node)

				// I don't fully understand when this happens, but
				// if we are unable to delete a marked node, move up
				// through the tree until we find something we can delete
				if c.Index() < 0 {
					toDelete[c.Parent()] = struct{}{}
					return true
				}
				codeDeleted = true
				c.Delete()
			}

			if _, ok := toDeleteParentType[node]; ok {
				delete(toDeleteParentType, node)
				if _, ok := node.(*dst.GenDecl); ok {
					if c.Index() < 0 {
						toDelete[c.Parent()] = struct{}{}
						return true
					}
					codeDeleted = true
					c.Delete()
				} else {
					toDeleteParentType[c.Parent()] = struct{}{}
				}
			}
			return true
		}).(*dst.File)

		outFiles[name] = newTree
	}

	return outFiles, codeDeleted, nil
}
