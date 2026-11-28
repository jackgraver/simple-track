// Command e2eseed empties the database named by DATABASE_URL and loads the e2e suite's baseline
// data (internal/e2e/seed). The e2e global setup runs it once to build the template database that
// every worker clones. With -profile it also adds that data profile (e.g. "large": two years of
// history, useful against a dev or perf database too). Refuses to run when APP_ENV is production.
//
// Usage: go run ./cmd/e2eseed [-profile large]   (from backend/)
package main

import (
	"errors"
	"flag"
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
	profile := flag.String("profile", "", "also load this data profile (large)")
	flag.Parse()
	apply, ok := seed.Profiles[*profile]
	if *profile != "" && !ok {
		return fmt.Errorf("unknown profile %q", *profile)
	}
	if err := env.Load(); err != nil {
		return err
	}
	if env.IsProduction() {
		return errors.New("refusing to run with APP_ENV=production")
	}
	db, err := database.ConnectOwner()
	if err != nil {
		return err
	}
	if err := seed.Reset(db); err != nil {
		return err
	}
	if apply != nil {
		if err := apply(db); err != nil {
			return fmt.Errorf("profile %s: %w", *profile, err)
		}
	}
	fmt.Println("e2eseed: baseline data loaded", *profile)
	return nil
}
