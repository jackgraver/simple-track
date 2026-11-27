package authrepo

import (
	"context"
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
	ctx := database.WithDB(context.Background(), db)

	expected := models.User{Username: "wanda", Password: "hash"}
	if err := CreateUser(ctx, &expected); err != nil {
		t.Fatal(err)
	}
	actual, err := FindUserByUsername(ctx, "wanda")
	if err != nil {
		t.Fatal(err)
	}
	if actual.ID != expected.ID || actual.Username != expected.Username {
		t.Fatalf("user: got %#v want %#v", actual, expected)
	}
}
