package commitlog

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	memCache "commitlog/cache"

	"golang.org/x/tools/cover"
)

type mockWriter struct {}
func (mw mockWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

type mockMemRunner struct {}
func (m mockMemRunner) GetCoverage(pkg string, test string) ([]*cover.Profile, error) {
	return []*cover.Profile{
		{
			FileName: pkg + "-" + test,
		},
	}, nil
}

type mockFileRunner struct {}
func (m mockFileRunner) GetCoverage(pkg string, test string) ([]*cover.Profile, error) {
	basePath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	coverFileName := path.Join(basePath, "testdata", fmt.Sprintf("coverage-%s.out", test))
	profiles, err := cover.ParseProfiles(coverFileName)
	if err != nil {
		return nil, err
	}

	return profiles, nil
}

func mockApp() commitlogApp {
	return commitlogApp{
		testRunner:        mockMemRunner{},
		testCoverageCache: memCache.New(),
		jobCache:          memCache.New(),
	}
}

func TestStartJob(t *testing.T) {
	app := mockApp()
	id := app.StartJob(JobConfig{
			tests: []string{"TestFuncOne"},
		sort: sortHardcodedOrder([]string{"TestFuncOne"}),
	})
	if id == "" {
		t.Errorf("expected non-empty id when starting job")
	}
}

func TestComputeFileContentsByTest(t *testing.T) {
	basePath, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	var (
		expectedFileContents [][]byte
		actualFileContents [][]byte
	)

	prunedFileNames := []string{"test_funcOne.go", "test_funcOneAndTwo.go", "test.go", "test.go"}
	for _, file := range prunedFileNames {
		f, err := os.Open(path.Join(basePath, "testdata", file))
		if err != nil {
			t.Fatal(err)
		}

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			t.Fatal(err)
		}
		expectedFileContents = append(expectedFileContents, contents)
	}

	expectedTestOrder := []string{"TestFuncOne", "TestFuncTwo", "TestFuncThree"}
	tests, files, err := computeFileContentsByTest(computationConfig{
		uuid:              "id-1",
		testCoverageCache: memCache.New(),
		statusWriter:      mockWriter{},
		runner:            mockFileRunner{},
		JobConfig:         JobConfig{
			pkg:   "testdata",
			tests: expectedTestOrder,
			sort:  sortHardcodedOrder(expectedTestOrder),
		},
	})

	if err != nil {
		t.Fatal("unexpected error: ", err)
	}

	if !reflect.DeepEqual(expectedTestOrder, tests) {
		t.Errorf("unexpected test ordering, got: %s, expected: %s", tests, expectedTestOrder)
	}

	for _, test := range files {
		for _, fileContents := range test {
			actualFileContents = append(actualFileContents, fileContents)
		}
	}

	for i := range actualFileContents {
		if strings.TrimSpace(string(expectedFileContents[i])) != strings.TrimSpace(string(actualFileContents[i])) {
			t.Errorf("unexpected file content %d, got: %s, expected: %s", i, actualFileContents[i], expectedFileContents[i])
		}
	}
}

func TestJobStatus(t *testing.T) {
	jobID := "id-1"
	app := commitlogApp{
		testRunner:        mockMemRunner{},
		testCoverageCache: memCache.New(),
		jobCache:          memCache.From(map[string]interface{}{
			jobID: jobCacheEntry{Complete: true},
		}),
	}
	status, err := app.JobStatus(jobID)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}
	if !status.Complete {
		t.Errorf("expected job to be complete")
	}
}

func TestGetTestProfiles_FallsBackToTestRunner(t *testing.T) {
	testCache := memCache.New()

	runnerCoverage, err := mockMemRunner{}.GetCoverage("pkg", "uncached")
	if err != nil {
		t.Errorf("unexpect error: %s", err)
	}

	profiles, err := getTestProfiles("pkg", "uncached", mockMemRunner{}, testCache)
	if err != nil {
		t.Errorf("unexpect error: %s", err)
	}
	if len(profiles) != len(runnerCoverage) {
		t.Errorf("Expected %d profiles got %d", len(runnerCoverage), len(profiles))
	}

	for i := range profiles {
		if profiles[i].FileName != runnerCoverage[i].FileName {
			t.Error("did not use test runner results")
		}
	}
}

func TestGetTestProfiles_UsesCache(t *testing.T) {
	testCache := memCache.New()
	cachedProfiles := []*cover.Profile{
		{
			FileName: "cache-filename",
		},
	}
	testCache.Write(cacheKeyForTest("pkg", "test1"), cachedProfiles)

	profiles, err := getTestProfiles("pkg", "test1", mockMemRunner{}, testCache)
	if err != nil {
		t.Errorf("unexpect error: %s", err)
	}
	if len(profiles) != len(cachedProfiles) {
		t.Errorf("Expected %d profiles got %d", len(cachedProfiles), len(profiles))
	}

	for i := range profiles {
		if profiles[i].FileName != cachedProfiles[i].FileName {
			t.Error("Did not use test cache")
		}
	}
}