// Command devuser ensures a local login exists for development.
//
// It creates (or resets the password of) DEV_AUTH_USER (default "dev") with
// password DEV_USER_PASSWORD (default "dev"). Idempotent; refuses to run when
// APP_ENV is production.
//
// Usage: go run ./cmd/devuser   (from backend/)
package main

import (
	"errors"
	"fmt"
	"os"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/env"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "devuser: %v\n", err)
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
	username := env.StringOr("DEV_AUTH_USER", "dev")
	password := env.StringOr("DEV_USER_PASSWORD", "dev")

	db, err := database.ConnectToPostgres()
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var user models.User
	err = db.Where("username = ?", username).First(&user).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		user = models.User{Username: username, Password: string(hash), Email: username + "@localhost"}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
		fmt.Printf("devuser: created %q (password %q)\n", username, password)
	case err != nil:
		return err
	default:
		if err := db.Model(&user).Update("password", string(hash)).Error; err != nil {
			return err
		}
		fmt.Printf("devuser: %q exists, password reset to %q\n", username, password)
	}
	return nil
}
