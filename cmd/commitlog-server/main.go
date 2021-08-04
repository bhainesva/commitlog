package main

import (
	"commitlog/cache"
	"log"
	"net/http"

	"commitlog"
	"commitlog/gocmd"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"golang.org/x/tools/cover"
)

type goPKGInfoProvider struct{}
func (g goPKGInfoProvider) ListPackages() ([]string, error) {
	return gocmd.List()
}
func (g goPKGInfoProvider) ListTests(pkg string) ([]string, error) {
	return gocmd.TestList(pkg)
}

type goTestRunner struct{}
func (g goTestRunner) GetCoverage(pkg string, test string) ([]*cover.Profile, error) {
	return gocmd.TestCover(pkg, test, "coverage.out")
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	commitLogApp := commitlog.NewCommitLogApp(
		goTestRunner{},
		cache.New(),
		cache.New(),
	)
	commitLogHandler := commitlog.Handler{
		Jobs: commitLogApp,
		LanguageInfo: goPKGInfoProvider{},
	}

	r.Get("/job/{id:[0-9a-zA-Z-]+}", commitLogHandler.JobStatus)
	r.Post("/job", commitLogHandler.StartJob)
	r.Post("/checkout", commitLogHandler.CheckoutFiles)
	r.Get("/listTests", commitLogHandler.Tests)
	r.Get("/listPackages", commitLogHandler.Packages)
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
