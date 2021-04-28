package main

import (
	"go/ast"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
	"log"
	"os"
)

func main() {
	removeDeadCode("./tmp")
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func removeDeadCode(pattern string) {
	conf := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
	}
	pkgs, err := packages.Load(conf, pattern)
	fatalIf(err)

	usedNames := map[string]struct{}{}
	unusedByName := map[string]ast.Node{}

	//usages := map[ast.Node]types.Object{}

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				if n, ok := n.(*ast.Ident); ok {
					if n.Name == "main" { return true }
					log.Println("Ident - Use Obj: ", n, pkg.TypesInfo.Uses[n])
					obj := pkg.Types.Scope().Lookup(n.Name)
					log.Println("Lookup scope: ", obj)

					if _, ok := pkg.TypesInfo.Uses[n]; ok {
						usedNames[n.Name] = struct{}{}
						delete(unusedByName, n.Name)
					} else {
						unusedByName[n.Name] = n
						return false
					}
				}
				return true
			})

			log.Println("unused names: ", unusedByName)

			toDelete := map[ast.Node]struct{}{}
			toDeleteParentType := map[ast.Node]struct{}{}
			newTree := astutil.Apply(file, func(c *astutil.Cursor) bool {
				if f, ok2 := c.Node().(*ast.Ident); ok2 {
					if unusedNode, ok := unusedByName[f.Name]; !ok || c.Node() != unusedNode {
						return true
					}

					parent := c.Parent()
					if p, ok := parent.(*ast.Field); ok {
						// Only one field declared, delete whole row
						if len(p.Names) == 1 {
							toDelete[parent] = struct{}{}
							return false
						} else { // delete just this identifier
							c.Delete()
							return false
						}
					}

					if _, ok := parent.(*ast.TypeSpec); ok {
						toDeleteParentType[parent] = struct{}{}
						return false
					}
				}

				return true
			}, func(c *astutil.Cursor) bool {
				if _, ok := toDelete[c.Node()]; ok {
					c.Delete()
				}

				if _, ok := toDeleteParentType[c.Node()]; ok {
					if _, ok := c.Node().(*ast.GenDecl); ok {
						c.Delete()
					} else {
						toDeleteParentType[c.Parent()] = struct{}{}
					}
				}
				return true
			})

			err := printer.Fprint(os.Stdout, token.NewFileSet(), newTree)
			fatalIf(err)
		}
	}

}
