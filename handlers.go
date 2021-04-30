package commitlog

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

func filterToTests(ss []string) []string {
	var out []string
	for _, s := range ss {
		if s != "" && len(strings.Fields(s)) == 1 && strings.HasPrefix(s, "Test") {
			out = append(out, s)
		}
	}

	return out
}

func HandlePackages(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("go", "list", "-find", "...")
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

func HandleTests(w http.ResponseWriter, r *http.Request) {
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

type filesRequest struct {
	Tests []string `json:"tests,omitempty"`
	Pkg   string   `json:"pkg,omitempty"`
	Sort  string   `json:"sort,omitempty"`
}

type checkoutRequest struct {
	Files map[string][]byte `json:"files,omitempty""`
}

type filesResponse struct {
	Tests []string            `json:"tests,omitempty"`
	Files []map[string][]byte `json:"files,omitempty""`
}

func HandleCheckoutFiles(w http.ResponseWriter, r *http.Request) {
	var req checkoutRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for fn, content := range req.Files {
		err := ioutil.WriteFile(fn, content, 0644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	return
}

func HandleTestFiles(w http.ResponseWriter, r *http.Request) {
	resp := filesResponse{
		Tests: []string{"Test1", "Test2", "Test3"},
		Files: []map[string][]byte{
			{
				"file/number/one.go":   []byte("Here is some file content"),
				"file/number/two.go":   []byte("Here is some file content"),
				"file/number/three.go": []byte("Here is some file content"),
				"file/number/four.go":  []byte("Here is some file content"),
			},
			{
				"file/number/one.go":   []byte("The content changes"),
				"file/number/two.go":   []byte("As we all do"),
				"file/number/three.go": []byte("oh no"),
				"file/number/four.go":  []byte("Big things ahead"),
			},
			{
				"file/number/one.go":   []byte("Now the files are happy"),
				"file/number/two.go":   []byte("This is their true form"),
				"file/number/three.go": []byte("I told you they could do it"),
				"file/number/four.go":  []byte("Yay for the files"),
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

func JobStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	outCh := make(chan Request)
	jobCacheChan <- Request{
		Type: READ,
		Key:  id,
		Out:  outCh,
	}

	info := <-outCh
	js, err := json.Marshal(info.Payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write(js)
	return
}

func HandleFiles(w http.ResponseWriter, r *http.Request) {
	var req filesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := uuid.New()
	jobCacheChan <- Request{
		Type: WRITE,
		Payload: jobCacheEntry{
			Complete: false,
			Details:  "Initializing job",
		},
		Key: id.String(),
	}

	w.Write([]byte(id.String()))

	go func() {
		results, err := jobOperation(id.String(), filesRequest{
			Tests: req.Tests,
			Pkg:   req.Pkg,
			Sort:  req.Sort,
		})
		if err != nil {
			jobCacheChan <- Request{
				Type: WRITE,
				Payload: jobCacheEntry{
					Complete: true,
					Details:  "",
					Error:    err,
				},
				Key: id.String(),
			}
		} else {
			jobCacheChan <- Request{
				Type: WRITE,
				Payload: jobCacheEntry{
					Complete: true,
					Details:  "",
					Error:    nil,
					Results:  results,
				},
				Key: id.String(),
			}
		}
	}()

	return
}

func jobOperation(id string, req filesRequest) (filesResponse, error) {
	var sortFunc testSortingFunction
	sortFunc = sortHardcodedOrder(req.Tests)
	if req.Sort == "raw" {
		sortFunc = sortTestsByRawLinesCovered
	} else if req.Sort == "net" {
		sortFunc = sortTestsByNewLinesCovered
	} else if req.Sort == "importance" {
		sortFunc = sortTestsByImportance
	}

	tests, fileContents, err := computeFileContentsByTest(computationConfig{
		pkg:   req.Pkg,
		tests: req.Tests,
		sort:  sortFunc,
		uuid:  id,
	})
	if err != nil {
		return filesResponse{}, err
	}

	return filesResponse{
		Tests: tests,
		Files: fileContents,
	}, nil
}
