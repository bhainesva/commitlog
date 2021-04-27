package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

var (
	// pkg -> test name -> coverage profiles
	coverageCache map[string]map[string][]*cover.Profile
)

func init() {
	coverageCache = map[string]map[string][]*cover.Profile{}
}

func main() {
	//run()
	//diff()
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hey buddy"))
	})

	r.Get("/listTests", handleTests)
	r.Post("/listFiles", handleFiles)
	r.Get("/listPackages", handlePackages)
	log.Println("Listening on port 3000...")
	http.ListenAndServe(":3000", r)
}

type filesRequest struct {
	Tests []string `json:"tests,omitempty"`
	Pkg   string   `json:"pkg,omitempty"`
}

func handleFiles(w http.ResponseWriter, r *http.Request) {
	var req filesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fileContents, err := runWithInfo(req.Pkg, req.Tests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	js, err := json.Marshal(fileContents)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(js)
}

func handleTests(w http.ResponseWriter, r *http.Request) {
	pkg := r.URL.Query().Get("pkg")
	cmd := exec.Command("go", "test", pkg, "-list", ".*")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Print(cmd.Stderr)
	}
	output := strings.Split(out.String(), "\n")
	tests := filterToTests(output)
	js, err := json.Marshal(tests)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Write(js)
}

func handlePackages(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("go", "list", "...")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Print(cmd.Stderr)
	}
	output := strings.Split(out.String(), "\n")
	js, err := json.Marshal(output)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	w.Write(js)
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func filterToTests(ss []string) []string {
	var out []string
	for _, s := range ss {
		if s != "" && len(strings.Fields(s)) == 1 && strings.HasPrefix(s, "Test") {
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

func mergeProfiles(profiles ...*cover.Profile) []*cover.Profile {
	type blockPos struct {
		SCol, ECol, SLine, ELine int
	}
	profileByFiles := map[string][]*cover.Profile{}

	for _, profile := range profiles {
		profileByFiles[profile.FileName] = append(profileByFiles[profile.FileName], profile)
	}

	var outProfiles []*cover.Profile

	for file, fileProfiles := range profileByFiles {
		blockByPos := map[blockPos]*cover.ProfileBlock{}
		for _, profile := range fileProfiles {
			for _, block := range profile.Blocks {
				block := block
				pos := blockPos{SCol: block.StartCol, ECol: block.EndCol, SLine: block.StartLine, ELine: block.EndLine}
				if existingBlock, ok := blockByPos[pos]; ok {
					if block.Count == 1 {
						existingBlock.Count = 1
					}
				} else {
					blockByPos[pos] = &block
				}
			}
		}

		var blocks []cover.ProfileBlock
		for _, block := range blockByPos {
			blocks = append(blocks, *block)
		}

		outProfiles = append(outProfiles, &cover.Profile{
			FileName: file,
			Blocks:   blocks,
		})
	}

	return outProfiles
}

func pre(fset *token.FileSet, profile *cover.Profile, m decorator.Map, deleteMap map[dst.Node]struct{}) func(cursor *dstutil.Cursor) bool {
	return func(cursor *dstutil.Cursor) bool {
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

			deleteMap[cursor.Parent()] = struct{}{}
			return true
		}
		return true
	}
}

func post(deleteMap map[dst.Node]struct{}) func(cursor *dstutil.Cursor) bool {
	return func(cursor *dstutil.Cursor) bool {
		if cursor.Node() == nil {
			return true
		}
		if _, ok := deleteMap[cursor.Node()]; ok {
			if cursor.Index() >= 0 {
				cursor.Delete()
				return true
			}
		}
		return true
	}
}

func getStrippedFiles(profiles []*cover.Profile) (map[string]*dst.File, error) {
	log.Println("Getting stripped files")
	files := map[string]*dst.File{}

	for _, profile := range profiles {
		tree, err := getStrippedFile(profile)
		if err != nil {
			return nil, err
		}
		files[profile.FileName] = tree
	}

	return files, nil
}

func getStrippedFile(profile *cover.Profile) (*dst.File, error) {
	log.Println("Finding file: ", profile.FileName)
	p, err := findFile(profile.FileName)
	if err != nil {
		return nil, err
	}
	log.Println("Found at: ", p)

	fSet := token.NewFileSet()
	d := decorator.NewDecorator(fSet)
	f, err := d.ParseFile(p, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	toDelete := map[dst.Node]struct{}{}
	newTree := dstutil.Apply(f, pre(fSet, profile, d.Map, toDelete), post(toDelete)).(*dst.File)
	return newTree, nil
}

func getTestProfile(pkg, test string) ([]*cover.Profile, error) {
	log.Println("Getting profile for: ", pkg, test)
	if testMap, ok := coverageCache[pkg]; ok {
		if profiles, ok := testMap[test]; ok {
			return profiles, nil
		}
	} else {
		coverageCache[pkg] = map[string][]*cover.Profile{}
	}
	cmd := exec.Command("go", "test", pkg, "-run", "^"+test+"$", "--coverprofile=coverage.out")
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	profiles, err := cover.ParseProfiles("coverage.out")
	if err != nil {
		return nil, err
	}
	os.Remove("coverage.out")

	coverageCache[pkg][test] = profiles
	return profiles, nil
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

func runWithInfo(pkg string, tests []string) ([]map[string][]byte, error) {
	out := make([]map[string][]byte, len(tests)+1)
	finalContentsMap := map[string][]byte{}

	log.Printf("Calculating diffs for %s, test order: %v", pkg, tests)

	var prevProfiles []*cover.Profile

	for i, test := range tests {
		log.Println("Adding test: ", test)
		profiles, err := getTestProfile(pkg, test)
		if err != nil {
			return nil, err
		}

		activeProfiles := mergeProfiles(append(prevProfiles, profiles...)...)

		contentsMap := map[string][]byte{}

		files, err := getStrippedFiles(activeProfiles)
		if err != nil {
			return nil, err
		}
		for name, tree := range files {
			var buf bytes.Buffer
			r := decorator.NewRestorer()
			err = r.Fprint(&buf, tree)
			if err != nil {
				return nil, err
			}

			contentsMap[name] = buf.Bytes()

			if _, ok := finalContentsMap[name]; !ok {
				fn, err := findFile(name)
				if err != nil {
					return nil, err
				}

				fullFileData, err := ioutil.ReadFile(fn)
				if err != nil {
					return nil, err
				}
				finalContentsMap[name] = fullFileData
			}
		}

		out[i] = contentsMap
		prevProfiles = activeProfiles
	}
	out[len(tests)] = finalContentsMap
	return out, nil
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

		files, err := getStrippedFiles(profiles)
		fatalIf(err)
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
