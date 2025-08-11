package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"

    "github.com/bugsnag/bugsnag-go"
    "github.com/gorilla/sessions"
    "github.com/kelseyhightower/envconfig"
    _ "github.com/lib/pq"
)

var store *sessions.CookieStore

const sessionName = "superwork"

var db *sql.DB

func main() {
    // Parse environment variables into config
    if err := envconfig.Process("superwork", &config); err != nil {
        log.Fatal(err)
    }

    // If running on a platform like Render with a provided PORT, use it
    if portStr := os.Getenv("PORT"); portStr != "" {
        if p, err := strconv.Atoi(portStr); err == nil {
            config.Port = p
        } else {
            log.Printf("WARNING: Invalid PORT value %q, using default port %d", portStr, config.Port)
        }
    }

    // Configure Bugsnag for error monitoring
    bugsnag.Configure(bugsnag.Configuration{
        APIKey: config.BugsnagAPIKey,
    })

    // Set up session cookie store
    store = sessions.NewCookieStore([]byte(config.Secret))

    // Validate OAuth2 redirect URLs (ensure they are set)
    if len(config.GoogleRedirect) == 0 {
        log.Panic("missing Google OAuth redirect URL")
    }
    if len(config.FacebookRedirect) == 0 {
        log.Panic("missing Facebook OAuth redirect URL")
    }

    // Set up logging to file if enabled
    if config.Log {
        f, err := os.OpenFile(config.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
        if err != nil {
            log.Fatalf("error opening log file: %v", err)
        }
        defer f.Close()
        log.SetOutput(f)
    }

    // Connect to the database (using DATABASE_URL or default local settings)
    if err := connectDB("superwork"); err != nil {
        log.Panic(err)
    }

    // Define all routes for the HTTP server
    r := defineRoutes()

    // Log the startup info
    log.Printf("App started on port %d", config.Port)

    // Start the HTTP server with Bugsnag error reporting
    log.Panic(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), bugsnag.Handler(r)))
}
