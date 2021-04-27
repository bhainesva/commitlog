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

	r.Get("/listTests", commitlog.HandleTests)
	r.Post("/listFiles", commitlog.HandleFiles)
	r.Get("/listPackages", commitlog.HandlePackages)
	log.Println("Listening on port 3000...")
	http.ListenAndServe(":3000", r)
}
