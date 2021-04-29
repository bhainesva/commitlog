package commitlog

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

func removeDeadCode(trees map[string]*dst.File, decorators map[string]*decorator.Decorator, pattern string) (map[string]*dst.File, error) {
	fullFileByShortFile := map[string]string{}
	conf := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		ParseFile: func(fset *token.FileSet, filename string, src []byte) (*ast.File, error) {
			d := decorators[filename]
			f := trees[filename]
			fullFileByShortFile[d.Ast.Nodes[f].(*ast.File).Name.Name] = filename
			return d.Ast.Nodes[f].(*ast.File), nil
		},
	}

	pkgs, err := packages.Load(conf, pattern)
	if err != nil {
		return nil, err
	}

	if len(pkgs) != 1 {
		return nil, fmt.Errorf("please only load 1 package")
	}
	pkg := pkgs[0]

	referencedTypePositions := map[token.Pos]struct{}{}
	deletionCandidates := map[token.Pos]struct{}{}

	for _, file := range pkg.Syntax {
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

				if usedObj, ok := pkg.TypesInfo.Uses[n]; ok {
					referencedTypePositions[n.Pos()] = struct{}{}
					delete(deletionCandidates, usedObj.Pos())
				} else {
					if _, ok := referencedTypePositions[n.Pos()]; !ok {
						deletionCandidates[n.Pos()] = struct{}{}
					}
				}
			}
			return true
		})
	}

	outFiles := map[string]*dst.File{}
	for _, file := range pkg.Syntax {
		// Sometimes we realize we want to delete a node while
		// visiting its children. These sets give information to the
		// post func to take care of that on the way back up the tree
		toDelete := map[dst.Node]struct{}{}
		toDeleteParentType := map[dst.Node]struct{}{}

		name := fullFileByShortFile[file.Name.Name]
		f := trees[name]
		d := decorators[name]
		// Traverse the AST a second time, deleting any identifiers
		// marked for deletion in the first pass
		newTree := dstutil.Apply(f, func(c *dstutil.Cursor) bool {
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
				c.Delete()
			}

			if _, ok := toDeleteParentType[node]; ok {
				delete(toDeleteParentType, node)
				if _, ok := node.(*dst.GenDecl); ok {
					if c.Index() < 0 {
						toDelete[c.Parent()] = struct{}{}
						return true
					}
					c.Delete()
				} else {
					toDeleteParentType[c.Parent()] = struct{}{}
				}
			}
			return true
		}).(*dst.File)

		outFiles[name] = newTree
	}

	return outFiles, nil
}
