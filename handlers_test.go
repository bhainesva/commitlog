package commitlog

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"commitlog/api"

	"github.com/go-chi/chi/v5"
)

type mockLanguageProvider struct{}
func (mlp mockLanguageProvider) ListPackages() ([]string, error) {
	return []string{"package-1", "package-2"}, nil
}
func (mlp mockLanguageProvider) ListTests(pkg string) ([]string, error) {
	return []string{"test-1", "test-2"}, nil
}

type mockJobManager struct {
	cache map[string]*jobCacheEntry
}
func (mjm mockJobManager) StartJob(conf JobConfig) string {
	return "id-1"
}
func (mjm mockJobManager) JobStatus(id string) (*jobCacheEntry, error) {
	return mjm.cache[id], nil
}

func TestPackagesHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := Handler{
		Jobs:         mockJobManager{},
		LanguageInfo: mockLanguageProvider{},
	}

	rr := httptest.NewRecorder()
	handler.Packages(rr, req)

	var response []string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("expected []string response, couldn't unmarshall: %s", rr.Body.String())
	}

	expectedPackages :=  []string{"package-1", "package-2"}
	if !reflect.DeepEqual(response, expectedPackages) {
		t.Errorf("unexpected response from Packages, got: %s, expected: %s", response, expectedPackages)
	}
}

func TestCheckoutFilesHandler(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.Remove(f.Name())
	}()

	fileContent := "content"
	checkoutRequest := api.CheckoutFilesRequest{Files: &api.FileMap{
		Files: map[string][]byte{f.Name(): []byte(fileContent)},
	}}
	bs, err := json.Marshal(checkoutRequest)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "", bytes.NewReader(bs))
	if err != nil {
		t.Fatal(err)
	}

	handler := Handler{
		Jobs:         mockJobManager{},
		LanguageInfo: mockLanguageProvider{},
	}

	rr := httptest.NewRecorder()
	handler.CheckoutFiles(rr, req)

	data, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Fatalf("failed to read checked out file: %s", err)
	}
	if string(data) != fileContent {
		t.Errorf("unexpected file content: %s", string(data))
	}
}

func TestTestsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	handler := Handler{
		Jobs:         mockJobManager{},
		LanguageInfo: mockLanguageProvider{},
	}

	rr := httptest.NewRecorder()
	handler.Tests(rr, req)

	var response []string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("expected []string response, couldn't unmarshall")
	}

	expectedTests :=  []string{"test-1", "test-2"}
	if !reflect.DeepEqual(response, expectedTests) {
		t.Errorf("unexpected response from Tests, got: %s, expected: %s", response, expectedTests)
	}
}

func TestStartJobHandler(t *testing.T) {
	jobRequest := api.StartJobRequest{}
	bs, err := json.Marshal(jobRequest)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "", bytes.NewReader(bs))
	if err != nil {
		t.Fatal(err)
	}

	jobManager := mockJobManager{cache: map[string]*jobCacheEntry{}}
	handler := Handler{
		Jobs:         jobManager,
		LanguageInfo: mockLanguageProvider{},
	}

	rr := httptest.NewRecorder()
	handler.StartJob(rr, req)

	var response api.StartJobResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("expected api.FetchFilesRequest response, couldn't unmarshall: %s", rr.Body.String())
	}

	if response.GetId() != "id-1" {
		t.Errorf("expected job id id-1, got: %s", response.GetId())
	}
}

func TestJobStatusHandler(t *testing.T) {
	req, err := http.NewRequest("POST", "", nil)
	if err != nil {
		t.Fatal(err)
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "id-1")

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	jobManager := mockJobManager{cache: map[string]*jobCacheEntry{
		"id-1":  {
			Complete: true,
			Results: jobResult{
				Tests: []string{},
				Files: []map[string][]byte{},
			},
		},
	}}
	handler := Handler{
		Jobs:         jobManager,
		LanguageInfo: mockLanguageProvider{},
	}

	rr := httptest.NewRecorder()
	handler.JobStatus(rr, req)

	var response api.JobStatusResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("expected api.FetchFilesRequest response, couldn't unmarshall: %s", rr.Body.String())
	}

	if !response.GetComplete() {
		t.Errorf("expected job to be complete, got incomplete")
	}
}

func TestCacheEntryToAPIResponse(t *testing.T) {
	type test struct {
		input *jobCacheEntry
		expectedOutput api.JobStatusResponse
	}
	tests := []test{
		{
			input: &jobCacheEntry{
				Complete: true,
			},
			expectedOutput: api.JobStatusResponse{
				Complete: true,
				Results: &api.JobResults{},
			},
		},
		{
			input: &jobCacheEntry{
				Details: "details",
			},
			expectedOutput: api.JobStatusResponse{
				Details: "details",
				Results: &api.JobResults{},
			},
		},
		{
			input: &jobCacheEntry{
				Error: "error",
			},
			expectedOutput: api.JobStatusResponse{
				Error: "error",
				Results: &api.JobResults{},
			},
		},
		{
			input: &jobCacheEntry{
				Results: jobResult{
					Tests: []string{"one", "two"},
					Files: []map[string][]byte{
						{"f1":[]byte("oneContent")},
						{"f2":[]byte("twoContent")},
					},
				},
			},
			expectedOutput: api.JobStatusResponse{
				Results: &api.JobResults{
					Tests: []string{"one", "two"},
					Files: []*api.FileMap{
						{Files: map[string][]byte{
							"f1":[]byte("oneContent"),
						}},
						{Files: map[string][]byte{
							"f2":[]byte("twoContent"),
						}},
					},
				},
			},
		},
	}

	for i, test := range tests {
		actualOutput := cacheEntryToAPIResponse(test.input)
		if !reflect.DeepEqual(actualOutput, test.expectedOutput) {
			log.Println(*actualOutput.Results)
			log.Println(*test.expectedOutput.Results)
			t.Errorf("case %d unexpected output, expected:\n%#v\ngot\n%#v", i, actualOutput, test.expectedOutput)
		}
	}
}

func TestRespondWithJson(t *testing.T) {
	type sampleStruct struct {
		String string `json:"string"`
		Int int `json:"int"`
	}
	inputStruct := sampleStruct{String: "str", Int: 4}
	jsonStr := `{"string":"str","int":4}`

	rr := httptest.NewRecorder()

	respondWithJSON(rr, inputStruct)

	if responseBody := rr.Body.String(); responseBody != jsonStr {
		t.Errorf("unexpected body %s, expected %s", responseBody, jsonStr)
	}
}

