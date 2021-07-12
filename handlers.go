package commitlog

import (
	"commitlog/api"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type languageProvider interface {
	List() ([]string, error)
	ListTests(pkg string) ([]string, error)
}

type jobManager interface {
	StartJob(jobConfig) string
	JobStatus(string) (*jobCacheEntry, error)
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

type Handler struct {
	Jobs jobManager
	LanguageInfo languageProvider
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

func (c *Handler) HandlePackages(w http.ResponseWriter, r *http.Request) {
	output, err := c.LanguageInfo.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, output)
}

func (c *Handler) HandleTests(w http.ResponseWriter, r *http.Request) {
	pkg := r.URL.Query().Get("pkg")
	tests, err := c.LanguageInfo.ListTests(pkg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, tests)
}

func (c *Handler) HandleCheckoutFiles(w http.ResponseWriter, r *http.Request) {
	var req api.CheckoutFilesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeFiles(req.GetFiles().GetFiles())
	return
}

func (c *Handler) JobStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	status, err := c.Jobs.JobStatus(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, cacheEntryToAPIResponse(status))
}

func (c *Handler) HandleFiles(w http.ResponseWriter, r *http.Request) {
	var req api.FetchFilesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := c.Jobs.StartJob(jobConfig{
		pkg:   req.GetPkg(),
		tests: req.GetTests(),
		sort:  req.GetSort(),
	})

	respondWithJSON(w, api.FetchFilesResponse{Id: id})
}
