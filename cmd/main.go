package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
)

func main() {
	pruneast("test.go")
	//r := chi.NewRouter()
	//r.Use(middleware.Logger)
	//r.Use(cors.Handler(cors.Options{
	//	AllowedOrigins: []string{"https://*", "http://*"},
	//	AllowedMethods: []string{"GET", "POST"},
	//	AllowedHeaders: []string{"Accept", "Content-Type"},
	//}))
	//
	//r.Get("/listTests", commitlog.HandleTests)
	//r.Post("/listFiles", commitlog.HandleFiles)
	//r.Get("/listPackages", commitlog.HandlePackages)
	//log.Println("Listening on port 3000...")
	//http.ListenAndServe(":3000", r)
}

type typeVisitor struct {}
func (v typeVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.Field:
		fmt.Println("Saw field: ", n.Type)
	case *ast.FieldList:
		fmt.Printf("Saw fieldlist: %#v\n", n.List)
	case *ast.Ident:
		if n.Obj != nil {
			fmt.Println("Saw ident: ", n.Name, n.Obj.Kind, n.Obj.Name, n.Obj.Data, n.Obj.Decl)
		} else {
			fmt.Println("Saw ident: ", n.Name)
		}
	case *ast.TypeSpec:
		fmt.Println("Saw typespec: ", n.Name)
	case *ast.StructType:
		fmt.Println("Saw structtype: ", n.Fields)
	}
	return v
}

func pruneast(fn string) {
	fSet := token.NewFileSet()
	//d := decorator.NewDecorator(fSet)
	_, err := parser.ParseFile(fSet, fn, nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
	}
	//ast.Walk(typeVisitor{}, f)

	conf := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(conf, "main")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pkgs)

	//used := make(map[types.Object]bool)
	for _, pkg := range pkgs {
		fmt.Println(pkg.TypesInfo)
		fmt.Println(pkg.TypesInfo.Uses)
	}
	//	for _, file := range pkg.Files {
	//		ast.Inspect(file, func(n ast.Node) bool {
	//			id, ok := n.(*ast.Ident)
	//			if !ok {
	//				return true
	//			}
	//			obj := pkg.Info.Uses[id]
	//			if obj != nil {
	//				used[obj] = true
	//			}
	//			return false
	//		})
	//	}
	//
	//	global := pkg.Pkg.Scope()
	//	var unused []types.Object
	//	for _, name := range global.Names() {
	//		if pkg.Pkg.Name() == "main" && name == "main" {
	//			continue
	//		}
	//		obj := global.Lookup(name)
	//		if !used[obj] && (pkg.Pkg.Name() == "main" || !ast.IsExported(name)) {
	//			unused = append(unused, obj)
	//		}
	//	}
	//	fmt.Println("UNUSED: ", unused)
	//}
}
