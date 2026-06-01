package database

import (
	"log"
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
	dsn := resolvePostgresDSN()
	
	const maxAttempts = 3
	const retryDelay = time.Second * 3 // 3s

	var db *gorm.DB
	var err error
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
			log.Printf("postgres failed connection: %d remaining", maxAttempts - attempt)
			time.Sleep(retryDelay)
		}
	}
	return nil, err
}

func resolvePostgresDSN() string {
	if d := env.OptionalString("DATABASE_URL"); d != "" {
		return d
	}
	appEnv := env.StringOr("APP_ENV", env.StringOr("GO_ENV", "development"))
	if appEnv == "prod" {
		return env.StringOr("DATABASE_URL_PRODUCTION", "postgres://postgres:postgres@localhost:5433/simpletracker_prod?sslmode=disable")
	}
	return env.StringOr("DATABASE_URL_DEVELOPMENT", "postgres://postgres:postgres@localhost:5432/simpletracker_dev?sslmode=disable")
}
