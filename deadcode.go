package commitlog

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"os"
)

func removeDeadCode(pattern string) error {
	conf := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(conf, pattern)
	if err != nil {
		return err
	}

	// Just testing on one package
	if len(pkgs) != 1 {
		return fmt.Errorf("please only load 1 package")
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

	for _, file := range pkg.Syntax {
		// Sometimes we realize we want to delete a node while
		// visiting its children. These sets give information to the
		// post func to take care of that on the way back up the tree
		toDelete := map[ast.Node]struct{}{}
		toDeleteParentType := map[ast.Node]struct{}{}

		// Traverse the AST a second time, deleting any identifiers
		// marked for deletion in the first pass
		newTree := astutil.Apply(file, func(c *astutil.Cursor) bool {
			if n, ok := c.Node().(*ast.Ident); ok {
				if _, ok := deletionCandidates[n.Pos()]; !ok {
					return true
				}

				parent := c.Parent()

				switch t := parent.(type) {
				case *ast.Field:
					// Only one field declared, delete whole row
					if len(t.Names) == 1 {
						toDelete[parent] = struct{}{}
						return false
					} else { // delete just this identifier
						c.Delete()
						return false
					}
				case *ast.TypeSpec:
					// In this case the node we want to delete isn't the direct
					// parent of the identifier. This struct is used to indicate
					// we should continue up the tree until we find a GenDecl
					toDeleteParentType[parent] = struct{}{}
					return false
				}
			}

			return true
		}, func(c *astutil.Cursor) bool {
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
				if _, ok := node.(*ast.GenDecl); ok {
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
		})

		err := printer.Fprint(os.Stdout, token.NewFileSet(), newTree)
		if err != nil {
			return err
		}
	}

	return nil
}
