package main

import (
	"commitlog"
	"commitlog/gocmd"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log"
	"net/http"
)

type goPkgInfo struct{}
func (g goPkgInfo) List() ([]string, error) {
	return gocmd.List()
}
func (g goPkgInfo) ListTests(pkg string) ([]string, error) {
	return gocmd.TestList(pkg)
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	commitLogApp := commitlog.NewCommitLogApp()
	commitLogHandler := commitlog.Handler{
		Jobs: commitLogApp,
		GoInfo: goPkgInfo{},
	}

	r.Get("/job/{id:[0-9a-zA-Z-]+}", commitLogHandler.JobStatus)
	r.Post("/checkout", commitLogHandler.HandleCheckoutFiles)
	r.Get("/listTests", commitLogHandler.HandleTests)
	r.Post("/listFiles", commitLogHandler.HandleFiles)
	r.Get("/listPackages", commitLogHandler.HandlePackages)
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
