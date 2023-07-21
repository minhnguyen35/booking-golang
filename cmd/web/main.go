package main

import (
	"encoding/gob"
	"flag"
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

	defer close(app.MailChan)
	listenForMail()

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
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})
	gob.Register(map[string]int{})


	inProduction := flag.Bool("production", true, "Application is in production")
	useCache := flag.Bool("cache", true, "Use cache template")
	dbName := flag.String("dbname", "", "Database name")
	dbHost := flag.String("dbhost", "localhost", "Database host")
	dbUser := flag.String("dbuser", "", "Database user")
	dbPW := flag.String("dbpass", "", "Database password")
	dbPort := flag.String("dbport", "5432", "Database port")
	dbSSL := flag.String("dbssl", "disable", "Database ssl settings (disable, prefer, require)")

	flag.Parse()
	if *dbName == "" || *dbUser == "" {
		fmt.Println("Missing required flags")
		os.Exit(1)
	}
	mailChan := make(chan models.MailData)
	app.MailChan = mailChan

	app.InProduction = *inProduction
	app.UseCache = *useCache

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
	connectionString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
			 *dbHost, *dbPort, *dbName, *dbUser, *dbPW, *dbSSL)
	db, err := driver.ConnectSQL(connectionString)
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
	repo := handlers.NewRepository(&app, db)
	render.NewRenderer(&app)
	helper.NewHelpers(&app)
	handlers.NewHandlers(repo)
	return db, nil
}
