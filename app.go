package commitlog

import (
	"bytes"
	"fmt"
	"github.com/dave/dst/decorator"
	"github.com/google/uuid"
	"golang.org/x/tools/cover"
	"io"
	"io/ioutil"
)

type commitlogApp struct {
	testRunner testRunner
	testCoverageCache cache
	jobCache          cache
}

type testRunner interface {
	// GetCoverage returns a go-style coverage profile given an identifier
	// for a package and test to run
	GetCoverage(pkg, test string) ([]*cover.Profile, error)
}

type JobConfig struct {
	pkg   string
	tests []string
	sort  testSortingFunction
}

type jobResult struct {
	Tests []string
	Files []map[string][]byte
}

type cache interface {
	Read(key string) interface{}
	Write(key string, value interface{})
	Delete(key string)
}

type jobCacheEntry struct {
	Complete bool
	Details  string
	Error    string
	Results  jobResult
}

func NewCommitLogApp(runner testRunner, testCoverageCache cache, jobCache cache) *commitlogApp {
	return &commitlogApp{
		testRunner: runner,
		testCoverageCache: testCoverageCache,
		jobCache:          jobCache,
	}
}

func (c *commitlogApp) StartJob(conf JobConfig) string {
	id := uuid.New()
	c.jobCache.Write(id.String(), jobCacheEntry{
		Complete: false,
		Details:  "Initializing job",
	})

	go func() {
		results, err := c.doJobOperation(id.String(), conf)
		if err != nil {
			c.jobCache.Write(id.String(), jobCacheEntry{
				Complete: true,
				Error:    err.Error(),
			})
		} else {
			c.jobCache.Write(id.String(), jobCacheEntry{
				Complete: true,
				Results:  results,
			})
		}
	}()

	return id.String()
}

func (c *commitlogApp) JobStatus(id string) (*jobCacheEntry, error) {
	info := c.jobCache.Read(id)
	if info == nil {
		return nil, nil
	}

	val, ok := info.(jobCacheEntry)
	if !ok {
		return nil, fmt.Errorf("unexpected type in job cache: %#v", val)
	}
	return &val, nil
}

type cacheWriter struct {
	id string
	cache
}
func (cw cacheWriter) Write(p []byte) (int, error) {
	cw.cache.Write(cw.id, jobCacheEntry{
		Complete: false,
		Details:  string(p),
	})
	return len(p), nil
}

func (c *commitlogApp) doJobOperation(id string, conf JobConfig) (jobResult, error) {
	tests, fileContents, err := computeFileContentsByTest(computationConfig{
		uuid:  id,
		testCoverageCache: c.testCoverageCache,
		statusWriter: cacheWriter{id: id, cache: c.jobCache},
		runner: c.testRunner,
		JobConfig: JobConfig{
			pkg:   conf.pkg,
			tests: conf.tests,
			sort:  conf.sort,
		},
	})
	if err != nil {
		return jobResult{}, err
	}

	return jobResult{
		Tests: tests,
		Files: fileContents,
	}, nil
}

type computationConfig struct {
	uuid  string
	testCoverageCache cache
	statusWriter io.Writer
	runner testRunner
	JobConfig
}

// computeFileContentsByTest calculates code diffs given a computationConfig.
// It returns the ordered tests, a map filename -> fileContents for each test, where the content
// is what is covered by the tests up to that point in the ordering, and an error.
func computeFileContentsByTest(config computationConfig) ([]string, []map[string][]byte, error) {
	var (
		pkg = config.pkg
		tests = config.tests
		prevProfiles []*cover.Profile
		out = make([]map[string][]byte, len(tests)+1)
		profilesByTest = testProfileData{}
		finalContentsMap = map[string][]byte{}
	)

	for i, test := range tests {
		config.statusWriter.Write([]byte(fmt.Sprintf("Computing coverage for %d of %d tests", i+1, len(tests))))
		profiles, err := getTestProfiles(pkg, test, config.runner, config.testCoverageCache)
		if err != nil {
			return nil, nil, err
		}
		config.testCoverageCache.Write(cacheKeyForTest(pkg, test), profiles)

		profilesByTest[test] = profiles
	}

	config.statusWriter.Write([]byte("Computing test ordering"))

	sortedTests := config.sort(profilesByTest)
	config.statusWriter.Write([]byte(fmt.Sprint("got sorted tests: ", sortedTests)))

	for i, test := range sortedTests {
		config.statusWriter.Write([]byte(fmt.Sprintf("Constructing diff %d of %d", i+1, len(sortedTests))))
		profiles := profilesByTest[test]
		activeProfiles, _ := mergeProfiles(prevProfiles, profiles)

		contentsMap := map[string][]byte{}

		config.statusWriter.Write( []byte("constructing dsts"))
		files, fset, ds, err := constructCoveredDSTs(activeProfiles, pkg)
		if err != nil {
			return nil, nil, err
		}

		config.statusWriter.Write([]byte("removing dead code"))
		// Parse package and kill dead code
		undeadFiles, updated, err := removeDeadCode(files, fset, ds)
		if err != nil {
			return nil, nil, err
		}
		for updated {
			undeadFiles, updated, err = removeDeadCode(files, fset, ds)
			if err != nil {
				return nil, nil, err
			}
		}

		config.statusWriter.Write([]byte("turning asts into []bytes"))
		// Convert ASTs into []byte
		for name, tree := range undeadFiles {
			var buf bytes.Buffer
			r := decorator.NewRestorer()
			err = r.Fprint(&buf, tree)
			if err != nil {
				return nil, nil, err
			}

			contentsMap[name] = buf.Bytes()

			if _, ok := finalContentsMap[name]; !ok {
				config.statusWriter.Write([]byte(fmt.Sprintf("loading final contents for %s", name)))
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

func getTestProfiles(pkg, test string, runner testRunner, testCache cache) ([]*cover.Profile, error) {
	info := testCache.Read(cacheKeyForTest(pkg, test))
	if info != nil {
		val, ok := info.([]*cover.Profile)
		if !ok {
			return nil, fmt.Errorf("unexpected type in test cache: %#v", val)
		}

		if val != nil {
			return val, nil
		}
	}

	profiles, err := runner.GetCoverage(pkg, test)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

func cacheKeyForTest(pkg, test string) string {
	return pkg + "-" + test
}