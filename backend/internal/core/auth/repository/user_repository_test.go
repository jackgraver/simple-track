package authrepo

import (
	"testing"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/database"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestUserRepositoryFindsUserByUsername(t *testing.T) {
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

	expected := models.User{Username: "wanda", Password: "hash"}
	if err := CreateUser(&expected); err != nil {
		t.Fatal(err)
	}
	actual, err := FindUserByUsername("wanda")
	if err != nil {
		t.Fatal(err)
	}
	if actual.ID != expected.ID || actual.Username != expected.Username {
		t.Fatalf("user: got %#v want %#v", actual, expected)
	}
}
