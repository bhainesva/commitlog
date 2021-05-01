package commitlog

import (
	"commitlog/api"
	"commitlog/gocmd"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type commitlogHandler struct {
	app commitlogApp
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

func NewCommitlogHandler(app commitlogApp) commitlogHandler {
	return commitlogHandler{
		app: app,
	}
}

func (c *commitlogHandler) HandlePackages(w http.ResponseWriter, r *http.Request) {
	output, err := gocmd.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, output)
}

func (c *commitlogHandler) HandleTests(w http.ResponseWriter, r *http.Request) {
	pkg := r.URL.Query().Get("pkg")
	tests, err := c.app.ListTests(pkg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, tests)
}

func (c *commitlogHandler) HandleCheckoutFiles(w http.ResponseWriter, r *http.Request) {
	var req api.CheckoutFilesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c.app.WriteFiles(req.GetFiles().GetFiles())
	return
}

func HandleTestFiles(w http.ResponseWriter, r *http.Request) {
	resp := api.JobResults{
		Tests: []string{"Test1", "Test2", "Test3"},
		Files: []*api.FileMap{
			{
				Files: map[string][]byte{
					"file/number/one.go":   []byte("Here is some file content"),
					"file/number/two.go":   []byte("Here is some file content"),
					"file/number/three.go": []byte("Here is some file content"),
					"file/number/four.go":  []byte("Here is some file content"),
				},
			},
			{
				Files: map[string][]byte{
					"file/number/one.go":   []byte("The content changes"),
					"file/number/two.go":   []byte("As we all do"),
					"file/number/three.go": []byte("oh no"),
					"file/number/four.go":  []byte("Big things ahead"),
				},
			},
			{
				Files: map[string][]byte{
					"file/number/one.go":   []byte("Now the files are happy"),
					"file/number/two.go":   []byte("This is their true form"),
					"file/number/three.go": []byte("I told you they could do it"),
					"file/number/four.go":  []byte("Yay for the files"),
				},
			},
		},
	}

	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(js)
}

func (c *commitlogHandler) JobStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	status, err := c.app.JobStatus(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, status)
}

func (c *commitlogHandler) HandleFiles(w http.ResponseWriter, r *http.Request) {
	var req api.FetchFilesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := c.app.StartJob(jobConfig{
		pkg:   req.GetPkg(),
		tests: req.GetTests(),
		sort:  req.GetSort(),
	})

	respondWithJSON(w, api.FetchFilesResponse{Id: id})
}
