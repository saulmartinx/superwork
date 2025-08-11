package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bugsnag/bugsnag-go"
	"github.com/gorilla/sessions"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
)

var store *sessions.CookieStore

const sessionName = "superwork"

var db *sql.DB

func main() {
	// Parse env variables
	if err := envconfig.Process("superwork", &config); err != nil {
		log.Fatal(err)
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey: config.BugsnagAPIKey,
	})

	// Setup cookie store
	store = sessions.NewCookieStore([]byte(config.Secret))

	// Validate oauth2 config
	if len(config.GoogleRedirect) == 0 {
		log.Panic("missing google oauth redirect url")
	}
	if len(config.FacebookRedirect) == 0 {
		log.Panic("missing facebook oauth redirect url")
	}

	// Open log file
	f, err := os.OpenFile(config.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	if config.Log {
		log.SetOutput(f)
	}

	if err := connectDB("superwork"); err != nil {
		log.Panic(err)
	}

	r := defineRoutes()

	log.Println("App started on http://localhost:" + fmt.Sprintf("%d", config.Port))

	// Serve HTTP
	log.Panic(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), bugsnag.Handler(r)))
}
