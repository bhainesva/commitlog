package commitlog

import (
	"encoding/json"
	"net/http"

	"commitlog/api"

	"github.com/go-chi/chi/v5"
)

type languageProvider interface {
	// ListPackages lists identifiers for units of code in a format that can be used
	// by ListTests to list tests
	// Only used for autocomplete so a default implementation can return an
	// empty list
	ListPackages() ([]string, error)
	// ListTests returns identifiers for tests associated with the given package
	ListTests(pkg string) ([]string, error)
}

type Handler struct {
	Jobs jobManager
	LanguageInfo languageProvider
}

type jobManager interface {
	// StartJob takes a JobConfig and returns a job identifier that can be
	// used to check the status of the job with JobStatus
	StartJob(JobConfig) string
	// JobStatus uses a job identifier to check the status of a job started with
	// StartJob. It does not return an error in the case the provided id isn't found
	JobStatus(string) (*jobCacheEntry, error)
}

// Packages responds to requests to list the available packages
func (c *Handler) Packages(w http.ResponseWriter, r *http.Request) {
	output, err := c.LanguageInfo.ListPackages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, output)
}

// Tests responds to requests to list the available tests for a package
// the requested package is provided through the `pkg` query param
func (c *Handler) Tests(w http.ResponseWriter, r *http.Request) {
	pkg := r.URL.Query().Get("pkg")
	tests, err := c.LanguageInfo.ListTests(pkg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, tests)
}

// CheckoutFiles writes the given file contents to disk
func (c *Handler) CheckoutFiles(w http.ResponseWriter, r *http.Request) {
	var req api.CheckoutFilesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeFiles(req.GetFiles().GetFiles())
	return
}

// JobStatus returns the status of the job requested using the `id`
// query parameter
func (c *Handler) JobStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	status, err := c.Jobs.JobStatus(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, cacheEntryToAPIResponse(status))
}

// StartJob begins processing a job according to the posted job config
func (c *Handler) StartJob(w http.ResponseWriter, r *http.Request) {
	var req api.StartJobRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sortFunc testSortingFunction
	switch req.GetSort() {
	case api.StartJobRequest_RAW:
		sortFunc = sortTestsByRawLinesCovered
	case api.StartJobRequest_NET:
		sortFunc = sortTestsByNewLinesCovered
	case api.StartJobRequest_IMPORTANCE:
		sortFunc = sortTestsByImportance
	case api.StartJobRequest_HARDCODED:
		sortFunc = sortHardcodedOrder(req.GetTests())
	}

	id := c.Jobs.StartJob(JobConfig{
		pkg:   req.GetPkg(),
		tests: req.GetTests(),
		sort:  sortFunc,
	})

	respondWithJSON(w, api.StartJobResponse{Id: id})
}

func respondWithJSON(w http.ResponseWriter, content interface{}) {
	js, err := json.Marshal(content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(js)
	return
}

func cacheEntryToAPIResponse(e *jobCacheEntry) api.JobStatusResponse {
	var filemaps []*api.FileMap
	for _, fm := range e.Results.Files {
		filemaps = append(filemaps, &api.FileMap{Files: fm})
	}

	return api.JobStatusResponse{
		Complete: e.Complete,
		Details:  e.Details,
		Error:    e.Error,
		Results: &api.JobResults{
			Tests: e.Results.Tests,
			Files: filemaps,
		},
	}
}

