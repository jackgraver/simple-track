package db

import (
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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