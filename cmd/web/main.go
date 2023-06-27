package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/minhnguyen/internal/config"
	"github.com/minhnguyen/internal/handlers"
	"github.com/minhnguyen/internal/models"
	"github.com/minhnguyen/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

// main is the main function
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	gob.Register(models.Reservation{})

	app.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		return err
	}

	app.TemplateCache = templateCache
	app.UseCache = false
	repo := handlers.NewRepository(&app)
	render.NewTemplates(&app)
	handlers.NewHandlers(repo)
	return nil
}
