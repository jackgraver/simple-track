package grocery

import (
	"be-simpletracker/internal/database"
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func groceryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&GroceryItem{}); err != nil {
		t.Fatal(err)
	}
	// Mirrors migration 0007's partial unique index.
	if err := db.Exec("CREATE UNIQUE INDEX uq_grocery_items_client_id ON grocery_items (client_id) WHERE client_id IS NOT NULL").Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCreateItemIdempotentByClientID(t *testing.T) {
	db := groceryTestDB(t)
	id := "11111111-1111-4111-8111-111111111111"
	a, err := CreateItem(database.WithDB(context.Background(), db), "milk", id)
	if err != nil {
		t.Fatal(err)
	}
	b, err := CreateItem(database.WithDB(context.Background(), db), "milk", id)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != b.ID {
		t.Fatalf("replay created a second row: %d vs %d", a.ID, b.ID)
	}
	var n int64
	db.Model(&GroceryItem{}).Count(&n)
	if n != 1 {
		t.Fatalf("want 1 row, got %d", n)
	}
	// Still returned after being completed.
	if _, err := CompleteItem(database.WithDB(context.Background(), db), itemRef{id: a.ID}); err != nil {
		t.Fatal(err)
	}
	c, err := CreateItem(database.WithDB(context.Background(), db), "milk", id)
	if err != nil || c.ID != a.ID {
		t.Fatalf("replay after complete: %v %v", c, err)
	}
}

func TestCreateItemWithoutClientIDAlwaysCreates(t *testing.T) {
	db := groceryTestDB(t)
	a, _ := CreateItem(database.WithDB(context.Background(), db), "eggs", "")
	b, _ := CreateItem(database.WithDB(context.Background(), db), "eggs", "")
	if a.ID == b.ID {
		t.Fatal("expected two rows")
	}
}
