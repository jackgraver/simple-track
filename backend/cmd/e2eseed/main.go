// Command e2eseed empties the database named by DATABASE_URL and loads the e2e suite's baseline
// data (internal/e2e/seed). The e2e global setup runs it once to build the template database that
// every worker clones. Refuses to run when APP_ENV is production.
//
// Usage: go run ./cmd/e2eseed   (from backend/)
package main

import (
	"errors"
	"fmt"
	"os"

	"be-simpletracker/internal/database"
	"be-simpletracker/internal/e2e/seed"
	"be-simpletracker/internal/env"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "e2eseed: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if err := env.Load(); err != nil {
		return err
	}
	if env.IsProduction() {
		return errors.New("refusing to run with APP_ENV=production")
	}
	db, err := database.ConnectToPostgres()
	if err != nil {
		return err
	}
	if err := seed.Reset(db); err != nil {
		return err
	}
	fmt.Println("e2eseed: baseline data loaded")
	return nil
}
