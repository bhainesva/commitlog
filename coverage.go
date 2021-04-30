package commitlog

import (
	"bytes"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"golang.org/x/tools/cover"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
)

// TODO: get rid of pkg global
var (
	// pkg -> test name -> coverage profiles
	coverageCache map[string]map[string][]*cover.Profile

	jobCacheChan chan Request
)

func init() {
	coverageCache = map[string]map[string][]*cover.Profile{}
	jobCacheChan = make(chan Request)
	go Cache(jobCacheChan)
}

type jobCacheEntry struct {
	Complete bool
	Details  string
	Error    error
	Results  filesResponse
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

func linesCovered(pp ...*cover.Profile) map[int]struct{} {
	lines := map[int]struct{}{}

	for _, p := range pp {
		for _, b := range p.Blocks {
			for line := b.StartLine; line <= b.EndLine; line++ {
				lines[line] = struct{}{}
			}
		}
	}

	return lines
}

func numLinesCovered(pp ...*cover.Profile) int {
	return len(linesCovered(pp...))
}

func sortTestsByRawLinesCovered(testProfiles map[string][]*cover.Profile) []string {
	var tests []string
	coverageByTest := map[string]int{}

	for test, profiles := range testProfiles {
		tests = append(tests, test)
		coverageByTest[test] = numLinesCovered(profiles...)
	}

	sort.Slice(tests, func(i, j int) bool {
		iCount := coverageByTest[tests[i]]
		jCount := coverageByTest[tests[j]]
		return iCount < jCount
	})

	return tests
}

func scoreLines(profiles []*cover.Profile) map[string]map[int]int {
	scores := map[string]map[int]int{}

	for _, profile := range profiles {
		if _, ok := scores[profile.FileName]; !ok {
			scores[profile.FileName] = map[int]int{}
		}

		for _, b := range profile.Blocks {
			if b.Count != 1 {
				continue
			}

			for line := b.StartLine; line <= b.EndLine; line++ {
				scores[profile.FileName][line]++
			}
		}
	}

	return scores
}

func scoreProfiles(profiles []*cover.Profile, lineWeights map[string]map[int]int) float64 {
	totalScore := 0.0
	totalLines := 0.0

	for _, p := range profiles {
		for _, b := range p.Blocks {
			for line := b.StartLine; line < b.EndLine; line++ {
				totalLines += 1
				totalScore += float64(lineWeights[p.FileName][line])
			}
		}
	}

	if totalLines == 0 {
		return 0
	}

	return totalScore / totalLines
}

// sortTestsByImportance sort tests using an 'importance' metric
// each line in a file is given a point for every test that covers it
// then tests are ranked by the average value of the lines they cover
func sortTestsByImportance(testProfiles map[string][]*cover.Profile) []string {
	allProfiles := []*cover.Profile{}
	tests := []string{}

	for test, ps := range testProfiles {
		tests = append(tests, test)
		allProfiles = append(allProfiles, ps...)
	}

	lineWeights := scoreLines(allProfiles)

	avgScoreByTest := map[string]float64{}
	for test, profiles := range testProfiles {
		avgScoreByTest[test] = scoreProfiles(profiles, lineWeights)
	}

	sort.Slice(tests, func(i, j int) bool {
		iCount := avgScoreByTest[tests[i]]
		jCount := avgScoreByTest[tests[j]]
		return iCount < jCount
	})

	return tests
}

func sortTestsByNewLinesCovered(testProfiles map[string][]*cover.Profile) []string {
	var sortedTests []string
	var tests []string
	var existingCoverage []*cover.Profile

	for test, _ := range testProfiles {
		tests = append(tests, test)
	}

	for len(tests) > 0 {
		minCoverageGain := -1
		minTestIdx := 0
		var minCoverage []*cover.Profile
		for i, test := range tests {
			profiles := testProfiles[test]
			newCoverage, coverageGain := mergeProfiles(existingCoverage, profiles)
			if minCoverageGain == 0 || coverageGain < minCoverageGain {
				minTestIdx = i
				minCoverageGain = coverageGain
				minCoverage = newCoverage
			}
			fmt.Println(test, " adds: ", coverageGain)
		}
		sortedTests = append(sortedTests, tests[minTestIdx])
		existingCoverage = minCoverage
		tests = append(tests[:minTestIdx], tests[minTestIdx+1:]...)
	}
	return sortedTests
}

func sortHardcodedOrder(order []string) testSortingFunction {
	return func(map[string][]*cover.Profile) []string {
		return order
	}
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

// TODO: the coverage calculation is suspicious here
// doesn't actually distinguish between existing and new profiles
// makes some assumptions that happen to be true
func mergeProfiles(existingProfiles, newProfiles []*cover.Profile) ([]*cover.Profile, int) {
	coverageGain := 0
	type blockPos struct {
		SCol, ECol, SLine, ELine int
	}
	profileByFiles := map[string][]*cover.Profile{}

	for _, profile := range append(existingProfiles, newProfiles...) {
		profileByFiles[profile.FileName] = append(profileByFiles[profile.FileName], profile)
	}

	var outProfiles []*cover.Profile

	for file, fileProfiles := range profileByFiles {
		blockByPos := map[blockPos]*cover.ProfileBlock{}
		for _, profile := range fileProfiles {
			profile := profile
			for _, block := range profile.Blocks {
				block := block
				pos := blockPos{SCol: block.StartCol, ECol: block.EndCol, SLine: block.StartLine, ELine: block.EndLine}
				if existingBlock, ok := blockByPos[pos]; ok {
					if block.Count == 1 {
						if existingBlock.Count == 1 {
							coverageGain -= block.EndLine - block.StartLine
						} else {
							existingBlock.Count = 1
						}
					}
				} else {
					blockByPos[pos] = &block
					if block.Count == 1 {
						coverageGain += block.EndLine - block.StartLine
					}
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

	return outProfiles, coverageGain
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

func getStrippedFiles(profiles []*cover.Profile) (map[string]*dst.File, *token.FileSet, map[string]*decorator.Decorator, error) {
	files := map[string]*dst.File{}
	decorators := map[string]*decorator.Decorator{}

	fset := token.NewFileSet()
	for _, profile := range profiles {
		tree, d, err := getStrippedFile(fset, profile)
		if err != nil {
			return nil, nil, nil, err
		}
		absfp, err := findFile(profile.FileName)
		if err != nil {
			return nil, nil, nil, err
		}
		files[absfp] = tree
		decorators[absfp] = d
	}

	return files, fset, decorators, nil
}

func getStrippedFile(fset *token.FileSet, profile *cover.Profile) (*dst.File, *decorator.Decorator, error) {
	p, err := findFile(profile.FileName)
	if err != nil {
		return nil, nil, err
	}

	d := decorator.NewDecorator(fset)
	f, err := d.ParseFile(p, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	toDelete := map[dst.Node]struct{}{}
	newTree := dstutil.Apply(f, pre(fset, profile, d.Map, toDelete), post(toDelete)).(*dst.File)
	return newTree, d, nil
}

func getTestProfile(pkg, test string) ([]*cover.Profile, error) {
	if testMap, ok := coverageCache[pkg]; ok {
		if profiles, ok := testMap[test]; ok {
			log.Println("Cache hit for: ", pkg, test)
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

type testSortingFunction func(map[string][]*cover.Profile) []string
type computationConfig struct {
	uuid  string
	pkg   string
	tests []string
	sort  testSortingFunction
}

// computeFileContentsByTest takes a package name and test ordering
// and returns a map filename -> fileContents for each test, where the content
// is what is covered by the tests up to that point in the ordering
func computeFileContentsByTest(config computationConfig) ([]string, []map[string][]byte, error) {
	pkg := config.pkg
	tests := config.tests

	out := make([]map[string][]byte, len(tests)+1)
	profilesByTest := map[string][]*cover.Profile{}
	finalContentsMap := map[string][]byte{}

	var prevProfiles []*cover.Profile

	for i, test := range tests {
		profiles, err := getTestProfile(pkg, test)
		jobCacheChan <- Request{
			Type: WRITE,
			Payload: jobCacheEntry{
				Complete: false,
				Details:  fmt.Sprintf("Computing coverage for %d of %d tests", i+1, len(tests)),
			},
			Key: config.uuid,
		}
		if err != nil {
			return nil, nil, err
		}

		profilesByTest[test] = profiles
	}

	jobCacheChan <- Request{
		Type: WRITE,
		Payload: jobCacheEntry{
			Complete: false,
			Details:  "Computing test ordering",
		},
		Key: config.uuid,
	}

	sortedTests := config.sort(profilesByTest)

	for i, test := range sortedTests {
		jobCacheChan <- Request{
			Type: WRITE,
			Payload: jobCacheEntry{
				Complete: false,
				Details:  fmt.Sprintf("Constructing diff %d of %d", i+1, len(sortedTests)),
			},
			Key: config.uuid,
		}
		profiles := profilesByTest[test]
		activeProfiles, _ := mergeProfiles(prevProfiles, profiles)

		contentsMap := map[string][]byte{}

		files, fset, ds, err := getStrippedFiles(activeProfiles)
		if err != nil {
			return nil, nil, err
		}

		// Parse package and kill dead code
		undeadFiles, updated, err := removeDeadCode(files, fset, ds)
		if err != nil {
			return nil, nil, err
		}
		for updated != false {
			undeadFiles, updated, err = removeDeadCode(files, fset, ds)
		}

		// Convert ASTs into []byte, get
		for name, tree := range undeadFiles {
			var buf bytes.Buffer
			r := decorator.NewRestorer()
			err = r.Fprint(&buf, tree)
			if err != nil {
				return nil, nil, err
			}

			contentsMap[name] = buf.Bytes()

			if _, ok := finalContentsMap[name]; !ok {
				fullFileData, err := ioutil.ReadFile(name)
				if err != nil {
					return nil, nil, err
				}
				finalContentsMap[name] = fullFileData
			}
		}

		out[i] = contentsMap
		prevProfiles = activeProfiles
	}
	out[len(tests)] = finalContentsMap
	return sortedTests, out, nil
}
