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

func removeDeadCode(trees map[string]*dst.File, fset *token.FileSet, decorators map[string]*decorator.Decorator) (map[string]*dst.File, bool, error) {
	codeDeleted := false
	conf := types.Config{
		Importer:         importer.Default(),
		IgnoreFuncBodies: false,
		Error:            func(error) {}, // Swallow errors
	}
	astByName := map[string]*ast.File{}
	var asts []*ast.File

	for fn, d := range decorators {
		astFile := d.Ast.Nodes[trees[fn]].(*ast.File)
		astByName[fn] = astFile
		asts = append(asts, astFile)
	}

	livingDSTNodes := map[dst.Node]struct{}{}
	livingPOS := map[token.Pos]struct{}{}
	for fn, tree := range trees {
		d := decorators[fn]
		dst.Inspect(tree, func(n dst.Node) bool {
			if n == nil {
				return true
			}

			ast := d.Ast.Nodes[n]

			livingDSTNodes[n] = struct{}{}
			livingPOS[ast.Pos()] = struct{}{}
			return true
		})
	}

	typesInfo := types.Info{
		Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object),
	}
	conf.Check("", fset, asts, &typesInfo)

	referencedTypePositions := map[token.Pos]struct{}{}
	deletionCandidates := map[token.Pos]struct{}{}

	for _, file := range astByName {
		// Inspect the AST, mark identifiers for deletion that
		// 1. Have a nil entry in TypesInfo.Uses
		//                  AND
		// 2. Are not themselves referenced by any other Ident node through
		//    TypesInfo.Uses
		ast.Inspect(file, func(n ast.Node) bool {
			if n, ok := n.(*ast.Ident); ok {
				if n.Name == "main" {
					return true
				}

				if usedObj, ok := typesInfo.Uses[n]; ok {
					if _, ok := livingPOS[n.Pos()]; ok {
						referencedTypePositions[usedObj.Pos()] = struct{}{}
						delete(deletionCandidates, usedObj.Pos())
					}
				} else {
					_, living := livingPOS[n.Pos()]
					if _, ok := referencedTypePositions[n.Pos()]; !ok || !living {
						deletionCandidates[n.Pos()] = struct{}{}
					}
				}
			}
			return true
		})
	}

	outFiles := map[string]*dst.File{}
	for name, file := range trees {
		// Sometimes we realize we want to delete a node while
		// visiting its children. These sets give information to the
		// post func to take care of that on the way back up the tree
		toDelete := map[dst.Node]struct{}{}
		toDeleteParentType := map[dst.Node]struct{}{}

		//f := trees[name]
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
