package database

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"be-simpletracker/internal/env"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Global database connection
var db_conn *gorm.DB

// Set the global database connection
func SetDB(db *gorm.DB) { db_conn = db }

// Get the global database connection
func GetDB() *gorm.DB { return db_conn }

// Connect to postgres
func ConnectToPostgres() (*gorm.DB, error) {
	if err := env.Load(); err != nil {
		return nil, err
	}
	// Get prod or dev dsn
	dsn, err := resolvePostgresDSN()
	if err != nil {
		return nil, err
	}

	const maxAttempts = 3
	const retryDelay = time.Second * 3 // 3s

	var db *gorm.DB
	// Try 3 times
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err == nil {
			db_conn = db
			log.Printf("postgres: connected successfully")
			return db, nil
		}
		// sleep to avoid hammering the database between attemps
		if attempt < maxAttempts {
			log.Printf("postgres failed connection: %d remaining", maxAttempts-attempt)
			time.Sleep(retryDelay)
		}
	}
	return nil, err
}

func resolvePostgresDSN() (string, error) {
	if d := env.OptionalString("DATABASE_URL"); d != "" {
		if env.IsProduction() {
			if err := validateProductionPostgresDSN(d); err != nil {
				return "", err
			}
		}
		return d, nil
	}
	if env.IsProduction() {
		d, err := env.String("DATABASE_URL_PRODUCTION")
		if err != nil {
			return "", fmt.Errorf("production database: %w", err)
		}
		if err := validateProductionPostgresDSN(d); err != nil {
			return "", err
		}
		return d, nil
	}
	return env.StringOr("DATABASE_URL_DEVELOPMENT", "postgres://postgres:postgres@localhost:5432/simpletracker_dev?sslmode=disable"), nil
}

func validateProductionPostgresDSN(dsn string) error {
	mode := postgresSSLMode(dsn)
	switch mode {
	case "require", "verify-ca", "verify-full":
		return nil
	default:
		return fmt.Errorf("production database must use sslmode=require, verify-ca, or verify-full")
	}
}

func postgresSSLMode(dsn string) string {
	if parsed, err := url.Parse(dsn); err == nil && (parsed.Scheme == "postgres" || parsed.Scheme == "postgresql") {
		return strings.ToLower(strings.TrimSpace(parsed.Query().Get("sslmode")))
	}
	for _, field := range strings.Fields(dsn) {
		key, value, ok := strings.Cut(field, "=")
		if ok && strings.EqualFold(key, "sslmode") {
			return strings.ToLower(strings.Trim(value, "'\""))
		}
	}
	return ""
}
