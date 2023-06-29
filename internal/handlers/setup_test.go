package handlers

import (
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	"github.com/minhnguyen/internal/config"
	"github.com/minhnguyen/internal/models"
	"github.com/minhnguyen/internal/render"
)

var testApp config.AppConfig
var session *scs.SessionManager
var pathToTemplate = "./../../templates"
var functions = template.FuncMap{}
var infoLog *log.Logger
var errorLog *log.Logger

func getRoutes() http.Handler{
	gob.Register(models.Reservation{})

	testApp.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate | log.Ltime)
	testApp.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	testApp.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = testApp.InProduction

	testApp.Session = session

	templateCache, err := CreateTestTemplateCache()
	if err != nil {
		return nil
	}

	testApp.TemplateCache = templateCache
	testApp.UseCache = true
	repo := NewRepository(&testApp)
	render.NewTemplates(&testApp)
	NewHandlers(repo)

	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	// mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", repo.Home)
	mux.Get("/about", repo.About)
	mux.Get("/general", repo.GeneralRoom)
	mux.Get("/contact", repo.Contact)
	mux.Get("/suite", repo.SuiteRoom)
	mux.Get("/search-availability", repo.SearchAvailability)
	mux.Post("/search-availability", repo.PostAvailability)
	mux.Post("/search-availability-json", repo.AvailabilityJSON)

	mux.Get("/make-reservation", Repo.Reservation)
	mux.Post("/make-reservation", Repo.PostReservation)
	mux.Get("/reservation-summary", Repo.ReservationSummary)

	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))
	return mux
}

func NoSurf(next http.Handler) http.Handler {
	csrf := nosurf.New(next)

	csrf.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path: "/",
		Secure: testApp.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrf
}

func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

func CreateTestTemplateCache() (map[string]*template.Template, error) {
	// templateCache := make(map[string]*template.Template)
	templateCache := map[string]*template.Template{}

	//get all template files
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplate))
	if err != nil {
		return templateCache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return templateCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
		if err != nil {
			return templateCache, err
		}
		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplate))
			if err != nil {
				return templateCache, err
			}
		}
		templateCache[name] = ts
	}
	return templateCache, nil
}
	