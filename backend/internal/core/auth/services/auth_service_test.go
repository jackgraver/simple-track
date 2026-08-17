package services

import (
	"errors"
	"testing"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/database"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatal(err)
	}
	database.SetDB(db)
	t.Cleanup(func() {
		database.SetDB(nil)
	})
}

func TestAuthServiceRegisterRedactsPassword(t *testing.T) {
	setupTestDB(t)
	service := NewAuthService(func(username string) (string, error) {
		return "token-" + username, nil
	})

	result, err := service.Register(RegisterInput{
		Username: "wanda",
		Password: "secret",
		Email:    "wanda@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Token != "token-wanda" {
		t.Fatalf("token: got %q", result.Token)
	}
	if result.User.Password != "" {
		t.Fatal("registered response contains password")
	}

	var stored models.User
	if err := database.GetDB().Where("username = ?", "wanda").First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Password == "" || stored.Password == "secret" {
		t.Fatal("stored password was not hashed")
	}
}

func TestAuthServiceRegisterRejectsDuplicateUsername(t *testing.T) {
	setupTestDB(t)
	if err := database.GetDB().Create(&models.User{Username: "wanda", Password: "hash"}).Error; err != nil {
		t.Fatal(err)
	}
	service := NewAuthService(func(string) (string, error) {
		return "token", nil
	})

	_, err := service.Register(RegisterInput{Username: "wanda", Password: "secret"})
	if !errors.Is(err, ErrUsernameExists) {
		t.Fatalf("error: got %v want %v", err, ErrUsernameExists)
	}
}

func TestAuthServiceLoginRejectsInvalidCredentials(t *testing.T) {
	setupTestDB(t)
	service := NewAuthService(func(string) (string, error) {
		return "token", nil
	})
	if _, err := service.Register(RegisterInput{Username: "wanda", Password: "secret"}); err != nil {
		t.Fatal(err)
	}

	_, err := service.Login(LoginInput{Username: "wanda", Password: "wrong"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("error: got %v want %v", err, ErrInvalidCredentials)
	}
}

func TestAuthServiceUpdatesBirthYear(t *testing.T) {
	setupTestDB(t)
	service := NewAuthService(func(string) (string, error) {
		return "token", nil
	})
	if _, err := service.Register(RegisterInput{Username: "wanda", Password: "secret"}); err != nil {
		t.Fatal(err)
	}
	user, err := service.UpdateBirthYear("wanda", 2000)
	if err != nil {
		t.Fatal(err)
	}
	if user.BirthYear == nil || *user.BirthYear != 2000 {
		t.Fatalf("birth year = %v, want 2000", user.BirthYear)
	}
}
