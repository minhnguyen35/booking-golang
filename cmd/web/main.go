package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/minhnguyen/internal/config"
	"github.com/minhnguyen/internal/driver"
	"github.com/minhnguyen/internal/handlers"
	"github.com/minhnguyen/internal/helper"
	"github.com/minhnguyen/internal/models"
	"github.com/minhnguyen/internal/render"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager
var infoLog *log.Logger
var errorLog *log.Logger

// main is the main function
func main() {
	db, err := run()
	if err != nil {
		log.Fatal(err)
	}

	defer db.SQL.Close()

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

func run() (*driver.DB, error) {
	gob.Register(models.Reservation{})
	gob.Register(models.Reservation{})
	gob.Register(models.Reservation{})
	gob.Register(models.Reservation{})

	app.InProduction = false

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	app.InfoLog = infoLog
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	app.ErrorLog = errorLog

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	//connect to db
	log.Println("Connecting to Database...")
	db, err := driver.ConnectSQL("host=localhost port=5432 dbname=booking user=minhnguyen password=123456")
	if err != nil {
		log.Fatal("Cannot connect to db! Dying...")
		return nil, err
	}
	log.Println("Connected to Database...")

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		return nil, err
	}

	app.TemplateCache = templateCache
	app.UseCache = false
	repo := handlers.NewRepository(&app, db)
	render.NewRenderer(&app)
	helper.NewHelpers(&app)
	handlers.NewHandlers(repo)
	return db, nil
}
