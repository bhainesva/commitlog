package commitlog

import (
	"bytes"
	"encoding/json"
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

type filesResponse struct {
	Tests []string            `json:"tests,omitempty"`
	Files []map[string][]byte `json:"files,omitempty""`
}

func HandleFiles(w http.ResponseWriter, r *http.Request) {
	var req filesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	js, err := json.Marshal(filesResponse{
		Tests: tests,
		Files: fileContents,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write(js)
}
