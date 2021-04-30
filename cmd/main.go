package main

import (
	"commitlog"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
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

	r.Get("/job/{id:[0-9a-zA-Z-]+}", commitlog.JobStatus)
	r.Post("/checkout", commitlog.HandleCheckoutFiles)
	r.Get("/listTests", commitlog.HandleTests)
	r.Post("/listFiles", commitlog.HandleFiles)
	r.Post("/listTestFiles", commitlog.HandleTestFiles)
	r.Get("/listPackages", commitlog.HandlePackages)
	http.ListenAndServe(":3000", r)
}
