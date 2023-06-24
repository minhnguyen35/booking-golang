package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/minhnguyen/pkg/config"
	"github.com/minhnguyen/pkg/handlers"
	"github.com/minhnguyen/pkg/render"
)

const portNumber = ":8080"
var app config.AppConfig
var session *scs.SessionManager

// main is the main function
func main() {

	app.InProduction = false

	session := scs.New()
	session.Lifetime = 24*time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction
	

	app.Session = session
	
	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache!")
	}

	app.TemplateCache = templateCache
	app.UseCache = true
	repo := handlers.NewRepository(&app)
	render.NewTemplates(&app)
	handlers.NewHandlers(repo)

	http.HandleFunc("/", handlers.Repo.Home)
	http.HandleFunc("/about", handlers.Repo.About)

	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))
	_ = http.ListenAndServe(portNumber, nil)
}
