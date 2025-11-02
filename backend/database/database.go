package database

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DefineRoutes(router *gin.Engine) {
	router.GET("/db/dump", func(c *gin.Context) {
		DumpSQLiteDB("st.db", "out_dump.sql")
		c.String(http.StatusOK, "Dump Successful")
	})

	router.GET("/db/restore", func(c *gin.Context) {
		RestoreSQLiteDB(c.MustGet("db").(*gorm.DB), "out_dump.sql")
		c.String(http.StatusOK, "Restore Successful")
	})
}

// Initializes a new database connection to a PostgreSQL database
func ConnectToPostgres() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=pass123 dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Initializes a new database connection to a SQLite database
func ConnectToSqlite() (*gorm.DB, error) {
    db, err := gorm.Open(sqlite.Open("st.db"), &gorm.Config{})
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

func RestoreSQLiteDB(db *gorm.DB, dumpPath string) error {
	content, err := ioutil.ReadFile(dumpPath)
	if err != nil {
		return fmt.Errorf("read dump failed: %w", err)
	}

	sql := string(content)

	// Execute without wrapping in transaction
	result := db.Session(&gorm.Session{SkipDefaultTransaction: true}).Exec(sql)
	return result.Error
}