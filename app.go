package commitlog

import (
	"bytes"
	"commitlog/cache"
	"fmt"
	"github.com/dave/dst/decorator"
	"github.com/google/uuid"
	"golang.org/x/tools/cover"
	"io/ioutil"
	"log"
)

type commitlogApp struct {
	testRunner testRunner
	testCoverageCache chan cache.Request
	jobCache          chan cache.Request
}

type jobConfig struct {
	pkg   string
	tests []string
	sort  string
}

type jobResult struct {
	Tests []string
	Files []map[string][]byte
}

type jobCacheEntry struct {
	Complete bool
	Details  string
	Error    string
	Results  jobResult
}

func (c *commitlogApp) writeJobCacheEntry(id string, entry jobCacheEntry) {
	cache.WriteEntry(c.jobCache, id, entry)
}

func (c *commitlogApp) finishJobWithError(id string, err error) {
	c.writeJobCacheEntry(id, jobCacheEntry{
		Complete: true,
		Error:    err.Error(),
	})
}

func (c *commitlogApp) updateInProgressJobStatus(id, details string) {
	c.writeJobCacheEntry(id, jobCacheEntry{
		Complete: false,
		Details:  details,
	})
}

func (c *commitlogApp) JobStatus(id string) (*jobCacheEntry, error) {
	outCh := make(chan cache.Request)
	c.jobCache <- cache.Request{
		Type: cache.READ,
		Key:  id,
		Out:  outCh,
	}
	info := <-outCh
	if info.Payload != nil {
		val, ok := info.Payload.(jobCacheEntry)
		if !ok {
			return nil, fmt.Errorf("unexpected payload type in job cache: %#v", val)
		}
		return &val, nil
	}

	return nil, nil
}

func (c *commitlogApp) StartJob(conf jobConfig) string {
	id := uuid.New()
	c.updateInProgressJobStatus(id.String(), "Initializing job")

	go func() {
		results, err := c.jobOperation(id.String(), conf)
		if err != nil {
			c.finishJobWithError(id.String(), err)
		} else {
			c.writeJobCacheEntry(id.String(), jobCacheEntry{
				Complete: true,
				Results:  results,
			})
		}
	}()

	return id.String()
}

type testRunner interface {
	GetCoverage(pkg, test string) ([]*cover.Profile, error)
}

func NewCommitLogApp(runner testRunner) *commitlogApp {
	testCoverageChannel := make(chan cache.Request)
	go cache.Initialize(testCoverageChannel)

	jobChannel := make(chan cache.Request)
	go cache.Initialize(jobChannel)

	return &commitlogApp{
		testRunner: runner,
		testCoverageCache: testCoverageChannel,
		jobCache:          jobChannel,
	}
}

func (c *commitlogApp) jobOperation(id string, conf jobConfig) (jobResult, error) {
	var sortFunc testSortingFunction
	sortFunc = sortHardcodedOrder(conf.tests)
	if conf.sort == "raw" {
		sortFunc = sortTestsByRawLinesCovered
	} else if conf.sort == "net" {
		sortFunc = sortTestsByNewLinesCovered
	} else if conf.sort == "importance" {
		sortFunc = sortTestsByImportance
	}

	tests, fileContents, err := c.computeFileContentsByTest(computationConfig{
		pkg:   conf.pkg,
		tests: conf.tests,
		sort:  sortFunc,
		uuid:  id,
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
	pkg   string
	tests []string
	sort  testSortingFunction
}

// computeFileContentsByTest takes a package name and test ordering
// and returns a map filename -> fileContents for each test, where the content
// is what is covered by the tests up to that point in the ordering
func (c *commitlogApp) computeFileContentsByTest(config computationConfig) ([]string, []map[string][]byte, error) {
	pkg := config.pkg
	tests := config.tests

	out := make([]map[string][]byte, len(tests)+1)
	profilesByTest := testProfileData{}
	finalContentsMap := map[string][]byte{}

	var prevProfiles []*cover.Profile

	for i, test := range tests {
		c.writeJobCacheEntry(config.uuid, jobCacheEntry{
			Complete: false,
			Details:  fmt.Sprintf("Computing coverage for %d of %d Tests", i+1, len(tests)),
		})
		profiles, err := c.getTestProfiles(pkg, test)
		if err != nil {
			return nil, nil, err
		}

		profilesByTest[test] = profiles
	}

	c.writeJobCacheEntry(config.uuid, jobCacheEntry{
		Complete: false,
		Details:  "Computing test ordering",
	})

	sortedTests := config.sort(profilesByTest)
	log.Println("got sorted tests: ", sortedTests)

	for i, test := range sortedTests {
		c.writeJobCacheEntry(config.uuid, jobCacheEntry{
			Complete: false,
			Details:  fmt.Sprintf("Constructing diff %d of %d", i+1, len(sortedTests)),
		})
		profiles := profilesByTest[test]
		activeProfiles, _ := mergeProfiles(prevProfiles, profiles)

		contentsMap := map[string][]byte{}

		log.Println("constructing dsts")
		files, fset, ds, err := constructCoveredDSTs(activeProfiles, pkg)
		if err != nil {
			return nil, nil, err
		}

		log.Println("removing dead code")
		// Parse package and kill dead code
		undeadFiles, updated, err := removeDeadCode(files, fset, ds)
		if err != nil {
			return nil, nil, err
		}
		for updated != false {
			undeadFiles, updated, err = removeDeadCode(files, fset, ds)
			if err != nil {
				return nil, nil, err
			}
		}

		log.Println("turning dsts to []bytes")
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
				log.Println("loading final contents for", name)
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

func (c *commitlogApp) getTestProfiles(pkg, test string) ([]*cover.Profile, error) {
	outCh := make(chan cache.Request)
	c.testCoverageCache <- cache.Request{
		Type: cache.READ,
		Key:  pkg + "-" + test,
		Out:  outCh,
	}
	info := <-outCh
	if info.Payload != nil {
		val, ok := info.Payload.([]*cover.Profile)
		if !ok {
			return nil, fmt.Errorf("unexpected payload type in test cache: %#v", val)
		}

		if val != nil {
			return val, nil
		}
	}

	profiles, err := c.testRunner.GetCoverage(pkg, test)
	if err != nil {
		return nil, err
	}

	cache.WriteEntry(c.testCoverageCache, pkg+"-"+test, profiles)
	return profiles, nil
}
