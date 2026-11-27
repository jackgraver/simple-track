package database

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/env"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Two database roles (docs/multi-user/DECISIONS.md, "App DB role"):
//
//   - the app role (AppRole, DATABASE_URL / DATABASE_URL_PRODUCTION) is what the server
//     connects as: DML only, no ownership, so row-level security applies to it.
//   - the owner role (DATABASE_OWNER_URL) owns the schema. cmd/migrate, devseed, e2eseed and
//     devuser use it. Without DATABASE_OWNER_URL they fall back to the app DSN (a single-role
//     setup) with a warning: migrations then fail at their first DDL, while DML-only tools
//     (seeds, devuser) still work.

// There is no global connection: the server hands its pool to Middleware, and request code
// reaches it only through From(ctx) (request.go).

// ConnectApp opens the server's pool on the app DSN.
func ConnectApp() (*gorm.DB, error) {
	return connect(AppDSN)
}

// ConnectOwner opens a pool on the owner DSN. For
// migrations, seeds and the other dev/ops tools, never the server.
func ConnectOwner() (*gorm.DB, error) {
	return connect(OwnerDSN)
}

func connect(dsnFn func() (string, error)) (*gorm.DB, error) {
	if err := env.Load(); err != nil {
		return nil, err
	}
	dsn, err := dsnFn()
	if err != nil {
		return nil, err
	}
	pool, err := PoolFromEnv()
	if err != nil {
		return nil, err
	}
	return Open(dsn, pool)
}

// Open connects to dsn (retrying briefly while Postgres starts) and applies pool.
func Open(dsn string, pool Pool) (*gorm.DB, error) {
	const maxAttempts = 3
	const retryDelay = time.Second * 3

	var err error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		var db *gorm.DB
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
			// created_at/updated_at follow the server clock (moved only by the e2e suite).
			NowFunc: func() time.Time { return clock.Now().Local() },
		})
		if err == nil {
			sqlDB, derr := db.DB()
			if derr != nil {
				return nil, derr
			}
			pool.apply(sqlDB)
			log.Printf("postgres: connected (%s)", pool)
			return db, nil
		}
		if attempt < maxAttempts {
			log.Printf("postgres failed connection: %d remaining", maxAttempts-attempt)
			time.Sleep(retryDelay)
		}
	}
	return nil, err
}

// AppDSN is the server's connection string: DATABASE_URL, else DATABASE_URL_PRODUCTION
// (production) or DATABASE_URL_DEVELOPMENT.
func AppDSN() (string, error) {
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

// OwnerDSN is the schema owner's connection string: DATABASE_OWNER_URL, else AppDSN.
func OwnerDSN() (string, error) {
	d := env.OptionalString("DATABASE_OWNER_URL")
	if d == "" {
		log.Printf("database: DATABASE_OWNER_URL not set; using the app DSN as the owner")
		return AppDSN()
	}
	if env.IsProduction() {
		if err := validateProductionPostgresDSN(d); err != nil {
			return "", fmt.Errorf("DATABASE_OWNER_URL: %w", err)
		}
	}
	return d, nil
}

func validateProductionPostgresDSN(dsn string) error {
	mode := postgresSSLMode(dsn)
	switch mode {
	case "require", "verify-ca", "verify-full":
		return nil
	case "disable":
		if strings.EqualFold(strings.TrimSpace(env.StringOr("DATABASE_ALLOW_INSECURE", "false")), "true") {
			return nil
		}
	default:
	}
	return fmt.Errorf("production database must use sslmode=require, verify-ca, or verify-full unless DATABASE_ALLOW_INSECURE=true")
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
