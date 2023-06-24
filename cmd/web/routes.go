package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minhnguyen/pkg/config"
	"github.com/minhnguyen/pkg/handlers"
)

func routes(app *config.AppConfig) http.Handler{
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/general", handlers.Repo.GeneralRoom)
	mux.Get("/contact", handlers.Repo.Contact)
	mux.Get("/suite", handlers.Repo.SuiteRoom)
	mux.Get("/search-availability", handlers.Repo.SearchAvailability)
	mux.Get("/make-reservation", handlers.Repo.Reservation)
	
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	
	return mux
}