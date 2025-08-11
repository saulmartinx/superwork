package main

import (
	"database/sql"
	"fmt"
    "os"
	"os/exec"
	"path/filepath"
)

func connectDB(dbname string) error {
	if db != nil {
		return nil
	}
    // Allow overriding the default Postgres connection string via an environment variable.
    // In production environments like Render or Fly.io, a DATABASE_URL (or similar) is
    // typically provided that contains the full DSN with host, user, password and dbname.
    dsn := os.Getenv("DATABASE_URL")
    if dsn == "" {
        // Fallback: attempt to read SUPERWORK_DATABASE_URL (legacy naming) if present.
        dsn = os.Getenv("SUPERWORK_DATABASE_URL")
    }
    if dsn == "" {
        // Default local connection string: use the provided dbname and assume a user
        // named "superwork" with no password and SSL disabled. This mirrors the
        // original behavior of the app when running locally with PostgreSQL.
        dsn = fmt.Sprintf("user=superwork dbname=%s sslmode=disable", dbname)
    }
    res, err := sql.Open("postgres", dsn)
    if err != nil {
        return err
    }
    db = res
    return nil
}

func recreateDB(dbname string) error {
	if _, err := exec.Command("dropdb", dbname).Output(); err != nil {
	}
	if _, err := exec.Command("dropuser", dbname).Output(); err != nil {
	}
	if _, err := exec.Command("createdb", dbname).Output(); err != nil {
		return err
	}
	if _, err := exec.Command("createuser", dbname).Output(); err != nil {
		return err
	}
	if _, err := exec.Command("psql", dbname, "-f", filepath.Join("db", "setup.sql")).CombinedOutput(); err != nil {
		return err
	}
	return nil
}
