package main

import (
	"bytes"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

var toDelete map[ast.Node]struct{}
var toDDelete map[dst.Node]struct{}

func init() {
	toDelete = map[ast.Node]struct{}{}
	toDDelete = map[dst.Node]struct{}{}
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func filterToTests(ss []string) []string {
	var out []string
	for _, s := range ss {
		if s != "" && len(strings.Fields(s)) == 1 {
			out = append(out, s)
		}
	}

	return out
}

func uncoveredLines(profile *cover.Profile) map[int]struct{} {
	lines := map[int]struct{}{}
	for _, block := range profile.Blocks {
		if block.Count == 0 {
			for i := block.StartLine; i <= block.EndLine; i++ {
				fmt.Println("Should skip: ", i)
				lines[i] = struct{}{}
			}
		}
	}

	return lines
}

func stripLines(lines []string, omit map[int]struct{}) []string {
	var out []string
	for i, line := range lines {
		if _, ok := omit[i+1]; !ok {
			out = append(out, line)
		}
	}

	return out
}

// golang.org/x/tools/cmd/cover/func.go
func findFile(file string) (string, error) {
	dir, file := filepath.Split(file)
	pkg, err := build.Import(dir, ".", build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("can't find %q: %v", file, err)
	}
	return filepath.Join(pkg.Dir, file), nil
}

func main() {
	//hmm()
	//comments()
	positions()
}

type commentVisitor struct {}
func (v commentVisitor) Visit(n dst.Node) dst.Visitor {
	if n == nil {
		return nil
	}
	decs := n.Decorations()
	fmt.Println("Node: ", decs.Start)
	return v
}

type positionVisitor struct {
	r *decorator.Restorer
	fset *token.FileSet
}
func (v positionVisitor) Visit(n dst.Node) dst.Visitor {
	if n == nil {
		return nil
	}
	aa := getAstByDst(v.r, n)
	if aa == nil {
		return v
	}

	log.Println(v.fset.PositionFor(aa.Pos(), true))
	printer.Fprint(os.Stdout, token.NewFileSet(), aa)
	fmt.Println("")
	return v
}

func positions() {
	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, "simple.go", nil, parser.ParseComments)
	fatalIf(err)


	r = decorator.NewRestorer()
	_, err = r.RestoreFile(f)
	fatalIf(err)
	dst.Walk(positionVisitor{r, fset}, f)
}



func comments() {
	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, "simple.go", nil, parser.ParseComments)
	fatalIf(err)

	done := false
	ntree := dstutil.Apply(f, func(c *dstutil.Cursor) bool {
		if c.Node() == nil || c.Node().Decorations() == nil {
			return true
		}
		if len([]string(c.Node().Decorations().Start)) > 0 {
			fmt.Println("Before: ", c.Node().Decorations().Start.All())
			if !done {
				c.Delete()
				done = true
			}
		}
		return true
	}, nil).(*dst.File)

	fi, err := os.Create("out.go")
	fatalIf(err)
	r := decorator.NewRestorer()
	r.Fprint(fi, ntree)

}

var r *decorator.Restorer

func inBadBlock(profile *cover.Profile, pos token.Position) bool {
	for _, block := range profile.Blocks {
		if block.Count != 0 {
			continue
		}

		if block.StartLine < pos.Line && pos.Line < block.EndLine {
			return true
		}

		if block.StartLine == pos.Line && block.EndLine == pos.Line {
			if block.StartCol <= pos.Column && pos.Column <= block.EndCol {
				return true
			}
		}

		if block.StartLine == pos.Line {
			if block.StartCol <= pos.Column {
				return true
			}
		}

		if block.EndLine == pos.Line {
			if pos.Column <= block.EndCol {
				return true
			}
		}
	}

	return false
}

func astpre(fset *token.FileSet, profile *cover.Profile) func (cursor *astutil.Cursor) bool {
	return func (cursor *astutil.Cursor) bool  {
		node := cursor.Node()
		if node == nil {
			return false
		}
		//fmt.Println("node: ", node)
		pos := node.Pos()
		position := fset.PositionFor(pos, false)
		if inBadBlock(profile, position) {
			if cursor.Index() >= 0 {
				fmt.Println("deleting: ", position)
				cursor.Delete()
				return true
			}
			fmt.Println("FAILING TO DELETE", pos)
			toDelete[cursor.Parent()] = struct{}{}
			return true
		}
		return true
	}
}

func astpost(cursor *astutil.Cursor) bool {
	if _, ok := toDelete[cursor.Node()]; ok {
		if cursor.Index() <= 1 {
			return true
		}
		fmt.Println("Cleaning up in post: ", cursor.Node().Pos())
		printer.Fprint(os.Stdout, token.NewFileSet(), cursor.Node())
		cursor.Delete()
	}
	return true
}

func getAstByDst(r *decorator.Restorer, node dst.Node) ast.Node {
	for a, d := range r.Ast.Nodes {
		if a == node {
			return d
		}
	}

	return nil
}

func pre(fset *token.FileSet, profile *cover.Profile, r *decorator.Restorer) func (cursor *dstutil.Cursor) bool {
	return func (cursor *dstutil.Cursor) bool  {
		node := cursor.Node()
		astNode := getAstByDst(r, node)
		parent := getAstByDst(r, cursor.Parent())
		if node == nil || astNode == nil {
			return false
		}
		pos := astNode.Pos()
		position := fset.PositionFor(pos, false)
		fmt.Println("IS ", position, " IN A BAD BLOCK: ", inBadBlock(profile, position))
		if inBadBlock(profile, position) {
			if cursor.Index() >= 0 {
				fmt.Println("DELETING")
				printer.Fprint(os.Stdout, token.NewFileSet(), astNode)
				fmt.Println("")
				//cursor.Delete()
				return false
			}

			fmt.Println("Cant delete: ")
			printer.Fprint(os.Stdout, token.NewFileSet(), astNode)
			fmt.Println("")
			fmt.Println("Queueing: ")
			printer.Fprint(os.Stdout, token.NewFileSet(), parent)
			fmt.Println("")
			toDDelete[cursor.Parent()] = struct{}{}
			return true
			//
			//return true
		}
		//fmt.Println("KEEPING")
		//printer.Fprint(os.Stdout, token.NewFileSet(), astNode)
		//fmt.Println("")
		return true
	}
}

func post(fset *token.FileSet, profile *cover.Profile, r *decorator.Restorer) func (cursor *dstutil.Cursor) bool {
	return  func(cursor *dstutil.Cursor) bool {
		if cursor.Node() == nil {
			return true
		}
		if _, ok := toDDelete[cursor.Node()]; ok{
			ugh := getAstByDst(r, cursor.Node())
			if cursor.Index() >= 0 {
				fmt.Println("Cleaning up")
				printer.Fprint(os.Stdout, token.NewFileSet(), ugh)
				fmt.Println("")
				//cursor.Delete()
				return true
			} else {
				fmt.Println("Cant fix")
				printer.Fprint(os.Stdout, token.NewFileSet(), ugh)
				fmt.Println("")
			}
		}
		return true
	}
}

func astff(profile *cover.Profile) ast.Node {
	p, _ := findFile(profile.FileName)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
	fatalIf(err)
	newtree := astutil.Apply(f, astpre(fset, profile), astpost)
	return newtree
}


func dave(profile *cover.Profile) {
	//p, _ := findFile("fmt/format.go")
	p := "simple.go"

	fset := token.NewFileSet()
	f, err := decorator.ParseFile(fset, p, nil, parser.ParseComments)
	fatalIf(err)

	r = decorator.NewRestorer()
	_, err = r.RestoreFile(f)
	fatalIf(err)

	newtree := dstutil.Apply(f, pre(fset, profile, r), post(fset, profile, r)).(*dst.File)
	fi, err := os.Create("out.go")

	r = decorator.NewRestorer()
	err = r.Fprint(fi, newtree)
	fatalIf(err)
}



func hmm() {
	pkg := "commitlog"
	cmd := exec.Command("go", "test", pkg, "-list", ".*")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	fatalIf(err)
	output := strings.Split(out.String(), "\n")
	tests := filterToTests(output)

	fmt.Println("testlist: ", tests)

	cmd = exec.Command("go", "test", pkg, "-run", tests[0], "--coverprofile=coverage.out")
	fmt.Println(cmd.Args)
	err = cmd.Run()
	fatalIf(err)

	fmt.Println("parsing profiles")
	profiles, err := cover.ParseProfiles("coverage.out")
	fatalIf(err)
	fmt.Println("filename: ", profiles[1].FileName)

	dave(profiles[1])

	//tree := astff(profiles[1])
	//f, err := os.Create("out.go")
	//fatalIf(err)
	//err = printer.Fprint(f, token.NewFileSet(), tree)
	//fatalIf(err)

	//toSkip := uncoveredLines(profiles[1])
	//
	//strippedFile := stripLines(strings.Split(string(file), "\n"), toSkip)
	//
	//err = ioutil.WriteFile("stripped.go", []byte(strings.Join(strippedFile, "\n")), 0644)
	//fatalIf(err)
}

