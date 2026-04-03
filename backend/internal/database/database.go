package database

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"be-simpletracker/internal/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Initializes a new database connection to a PostgreSQL database.
func ConnectToPostgres() (*gorm.DB, error) {
	utils.LoadEnvIfNeeded()

	dsn := resolvePostgresDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func resolvePostgresDSN() string {
	if dsn, ok := os.LookupEnv("DATABASE_URL"); ok && strings.TrimSpace(dsn) != "" {
		return dsn
	}

	appEnv := normalizeEnv(getEnv("APP_ENV", getEnv("GO_ENV", "development")))
	if appEnv == "production" {
		return getEnv("DATABASE_URL_PRODUCTION", "postgres://postgres:postgres@localhost:5433/simpletracker_prod?sslmode=disable")
	}

	return getEnv("DATABASE_URL_DEVELOPMENT", "postgres://postgres:postgres@localhost:5432/simpletracker_dev?sslmode=disable")
}

func normalizeEnv(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "prod", "production":
		return "production"
	default:
		return "development"
	}
}

// Initializes a new database connection to a SQLite database
func ConnectToSqlite(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func DumpSQLiteDB(dbPath string, dumpPath string) error {
	cmd := exec.Command("sqlite3", dbPath, ".dump")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dump command failed: %w", err)
	}

	lines := []string{}
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading dump failed: %w", err)
	}

	cleaned := []string{}
	for i, line := range lines {
		if i == 1 && strings.Contains(strings.ToUpper(line), "TRANSACTION") {
			continue
		}
		cleaned = append(cleaned, line)
	}

	for i := len(cleaned) - 1; i >= 0; i-- {
		if strings.TrimSpace(cleaned[i]) != "" {
			if strings.Contains(strings.ToUpper(cleaned[i]), "COMMIT") {
				cleaned = append(cleaned[:i], cleaned[i+1:]...)
			}
			break
		}
	}

	return os.WriteFile(dumpPath, []byte(strings.Join(cleaned, "\n")), 0644)
}

func RestoreSQLiteDB(dumpPath string) error {
	conn, err := ConnectToSqlite("st.db")
	if err != nil {
		return err
	}
	inst, _ := conn.DB()
	defer inst.Close()

	sqlBytes, err := os.ReadFile("out_dump.sql")
	if err != nil {
		log.Fatal(err)
	}

	sql := string(sqlBytes)

	result := conn.Session(&gorm.Session{SkipDefaultTransaction: true}).Exec(sql)
	return result.Error
}
