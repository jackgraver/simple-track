package database

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func DefineRoutes(router *gin.Engine) {
	router.POST("/db/dump", func(c *gin.Context) {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := fmt.Sprintf("database/dumps/dbdump_%s.sql", timestamp)
		dbPath := getEnv("DB_PATH", "st.db")

		if err := DumpSQLiteDB(dbPath, filename); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Dump failed: %v", err))
			return
		}

		c.String(http.StatusOK, fmt.Sprintf("Dump successful: %s", filename))
	})

	router.POST("/db/restore", func(c *gin.Context) {
		RestoreSQLiteDB("out_dump.sql")
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

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Initializes a new database connection to a SQLite database
func ConnectToSqlite() (*gorm.DB, error) {
	dbPath := getEnv("DB_PATH", "st.db")
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
	conn, err := ConnectToSqlite()
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