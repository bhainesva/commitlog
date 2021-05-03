package main

import (
	"commitlog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Accept", "Content-Type"},
	}))

	commitLogApp := commitlog.NewCommitLogApp()
	commitLogHandler := commitlog.NewCommitlogHandler(commitLogApp)

	r.Get("/job/{id:[0-9a-zA-Z-]+}", commitLogHandler.JobStatus)
	r.Post("/checkout", commitLogHandler.HandleCheckoutFiles)
	r.Get("/listTests", commitLogHandler.HandleTests)
	r.Post("/listFiles", commitLogHandler.HandleFiles)
	r.Post("/listTestFiles", commitlog.HandleTestFiles)
	r.Get("/listPackages", commitLogHandler.HandlePackages)
	err := http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatal(err)
	}
}
