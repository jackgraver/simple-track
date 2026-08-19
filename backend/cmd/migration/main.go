package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Migrations struct {
	version   string
	appliedAt time.Time
}

func main() {
	// db, err := database.ConnectToPostgres()
	// if err != nil {
	// 	panic("Database unreachable")
	// }

	// // fetch the last migration and ensure table exists
	// var lastMigration Migrations

	// db.Last(&lastMigration)
	// // db.Raw("SELECT * FROM migrations ORDER BY applied_at").Scan(&lastMigration)

	// fmt.Println(lastMigration)

	// // if lastMigration.version == "" {
	// // 	panic("No migrations table found")
	// // }

	args := os.Args[1:]

	migrationsPath := "."
	if len(args) >= 1 {
		migrationsPath = args[0]
	}

	migrations, err := os.ReadDir(migrationsPath)
	if err != nil {
		panic("Failed to read migrations directory")
	}

	if !CheckIfInitialized(migrationsPath) {
		return
	}

	for _, migration := range migrations {
		if len(migration.Name()) <= 3 {
			continue
		}
		migrationNum := migration.Name()[0:3]

		// Dont run init migration
		if migrationNum != "000" {
			res, err := os.ReadFile(filepath.Join(migrationsPath, migration.Name()))
			if err != nil {
				panic("Failed to read migration file")
			}

			fmt.Println(string(res))
		}
	}

}

var initText = `-- Init creates the migrations table
-- v.000
-- <date>

CREATE TABLE IF NOT EXISTS migrations (
    version VARCHAR(15) PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

// Init creates the migrations table
// It writes the initial migration file to the migrations directory
func CheckIfInitialized(migrationsPath string) bool {
	// Get all migration files in migration directory
	migrations, err := os.ReadDir(migrationsPath)
	if err != nil {
		panic("Failed to read migrations directory")
	}

	// if there are migration files the migration system is initialized
	if len(migrations) > 0 {
		return true
	}

	fmt.Println("[INFO] Initializing migration system...")

	// Insert current date into init file comments
	initText = strings.Replace(initText, "<date>", time.Now().Format("01/02/2006"), 1)

	iB := []byte(initText)
	migPath := filepath.Join(migrationsPath, "000_init.sql")
	err = os.WriteFile(migPath, iB, 0644)
	if err != nil {
		panic(err)
	}

	// Return system was not initialized
	return false
}

var newMigText = `--
-- v.<version>
-- <date>
`

// NewMigration creates a new migration file
// It writes the migration text to the migrations directory
func NewMigration(previousVerion string, migrationsPath string) {
	newMigText = strings.Replace(newMigText, "<version>", IncrementVersion(previousVerion), 1)
	newMigText = strings.Replace(newMigText, "<date>", time.Now().Format("01/02/2006"), 1)

	iB := []byte(newMigText)
	migPath := filepath.Join(migrationsPath, IncrementVersion(previousVerion)+"_banada.sql")
	err := os.WriteFile(migPath, iB, 0644)
	if err != nil {
		panic(err)
	}
}

func IncrementVersion(version string) string {
	versionInt, _ := strconv.Atoi(version)
	versionInt++
	return fmt.Sprintf("%03d", versionInt)
}
