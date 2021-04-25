package main

import (
	"bytes"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sergi/go-diff/diffmatchpatch"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

var toDDelete map[dst.Node]struct{}

func init() {
	toDDelete = map[dst.Node]struct{}{}
}

func main() {
	//run()
	//diff()
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hey buddy"))
	})
	r.Get("/listTests/{pkg}", handleTests)
	http.ListenAndServe(":3000", r)
}

func handleTests(w http.ResponseWriter, r *http.Request) {
	pkg := chi.URLParam(r, "pkg")
	cmd := exec.Command("go", "test", pkg, "-list", ".*")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	fatalIf(err)
	output := strings.Split(out.String(), "\n")
	tests := filterToTests(output)

	w.Write([]byte(strings.Join(tests, ", ")))
}


func diff() {
	text1 := "Lorem ipsum dolor."
	text2 := "Lorem dolor sit amet."

	dmp := diffmatchpatch.New()

	diffs := dmp.DiffMain(text1, text2, false)

	fmt.Println(dmp.DiffPrettyText(diffs))
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

// golang.org/x/tools/cmd/cover/func.go
func findFile(file string) (string, error) {
	dir, file := filepath.Split(file)
	pkg, err := build.Import(dir, ".", build.FindOnly)
	if err != nil {
		return "", fmt.Errorf("can't find %q: %v", file, err)
	}
	return filepath.Join(pkg.Dir, file), nil
}

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

func getAstByDst(m decorator.Map, node dst.Node) ast.Node {
	for a, d := range m.Ast.Nodes {
		if a == node {
			return d
		}
	}

	return nil
}

func pre(fset *token.FileSet, profile *cover.Profile, m decorator.Map) func (cursor *dstutil.Cursor) bool {
	return func (cursor *dstutil.Cursor) bool  {
		node := cursor.Node()
		astNode := getAstByDst(m, node)
		if node == nil || astNode == nil {
			return false
		}
		pos := astNode.Pos()
		position := fset.PositionFor(pos, false)
		if inBadBlock(profile, position) {
			if cursor.Index() >= 0 {
				cursor.Delete()
				return false
			}

			toDDelete[cursor.Parent()] = struct{}{}
			return true
		}
		return true
	}
}

func post (cursor *dstutil.Cursor) bool {
	if cursor.Node() == nil {
		return true
	}
	if _, ok := toDDelete[cursor.Node()]; ok{
		if cursor.Index() >= 0 {
			cursor.Delete()
			return true
		}
	}
	return true
}

func getStrippedFiles(profiles []*cover.Profile) map[string]*dst.File {
	files := map[string]*dst.File{}

	for _, profile := range profiles {
		tree := getStrippedFile(profile)
		files[profile.FileName] = tree
	}

	return files
}

func getStrippedFile(profile *cover.Profile) *dst.File {
	p, err := findFile(profile.FileName)
	fatalIf(err)

	fset := token.NewFileSet()
	d := decorator.NewDecorator(fset)
	f, err := d.ParseFile( p, nil, parser.ParseComments)
	fatalIf(err)

	newtree := dstutil.Apply(f, pre(fset, profile, d.Map), post).(*dst.File)
	return newtree
}

func generateProfiles(pkg string, tests []string) ([]*cover.Profile, error) {
	cmd := exec.Command("go", "test", pkg, "-run", strings.Join(tests, "|"), "--coverprofile=coverage.out")
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	profiles, err := cover.ParseProfiles("coverage.out")
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

func run() {
	//pkg := "fmt"
	pkg := "commitlog/simple"
	cmd := exec.Command("go", "test", pkg, "-list", ".*")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	fatalIf(err)
	output := strings.Split(out.String(), "\n")
	tests := filterToTests(output)

	fmt.Println("Available Rests: ", tests)
	activeTests := []string{}
	previousFileVersion := ""

	for _, test := range tests {
		fmt.Println("Adding test: ", test)
		fmt.Println("------------------------")
		activeTests = append(activeTests, test)
		profiles, err := generateProfiles(pkg, activeTests)
		fatalIf(err)

		files := getStrippedFiles(profiles)
		for name, tree := range files {
			if name != "commitlog/simple/simple.go" {
				continue
			}

			var buf bytes.Buffer
			dmp := diffmatchpatch.New()
			r := decorator.NewRestorer()
			err = r.Fprint(&buf, tree)

			newVersion := buf.String()
			diffs := dmp.DiffMain(previousFileVersion, newVersion, false)
			fmt.Println(dmp.DiffPrettyText(diffs))

			previousFileVersion = newVersion
		}
	}

	fullFileData, err := ioutil.ReadFile("../simple/simple.go")
	fatalIf(err)
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(previousFileVersion, string(fullFileData), false)
	if len(diffs) != 0 {
		fmt.Println("Uncovered code ----------")
		fmt.Println(dmp.DiffPrettyText(diffs))
	}
}

