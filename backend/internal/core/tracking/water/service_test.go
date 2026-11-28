package water

import (
	"be-simpletracker/internal/database"
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// asUser is a context for module tests on SQLite: the DB plus a signed-in user (there is no RLS here;
// the two-user Postgres tests cover isolation).
func asUser(db *gorm.DB, id uint) context.Context {
	return database.WithIdentity(database.WithDB(context.Background(), db), database.Identity{UserID: id})
}

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&DrinkSizePreset{}, &WaterLog{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX uq_water_logs_user_client_id ON water_logs(user_id, client_id) WHERE client_id IS NOT NULL").Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestCreateWaterLogIdempotentByClientID(t *testing.T) {
	db := testDB(t)
	day := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	id := "6f1c2c1e-0000-4000-8000-000000000001"

	a, err := CreateWaterLog(asUser(db, 1), day, 16, nil, id)
	if err != nil {
		t.Fatal(err)
	}
	b, err := CreateWaterLog(asUser(db, 1), day, 16, nil, id)
	if err != nil {
		t.Fatal(err)
	}
	if a.ID != b.ID {
		t.Fatalf("replay created a second row: %d vs %d", a.ID, b.ID)
	}
	var n int64
	db.Model(&WaterLog{}).Count(&n)
	if n != 1 {
		t.Fatalf("want 1 row, got %d", n)
	}

	// Replay after delete does not resurrect or duplicate.
	if err := DeleteWaterLog(asUser(db, 1), a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateWaterLog(asUser(db, 1), day, 16, nil, id); err != nil {
		t.Fatal(err)
	}
	db.Model(&WaterLog{}).Count(&n)
	if n != 0 {
		t.Fatalf("deleted log came back: %d live rows", n)
	}
}

func TestCreateWaterLogWithoutClientIDAlwaysCreates(t *testing.T) {
	db := testDB(t)
	day := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 2; i++ {
		if _, err := CreateWaterLog(asUser(db, 1), day, 8, nil, ""); err != nil {
			t.Fatal(err)
		}
	}
	var n int64
	db.Model(&WaterLog{}).Count(&n)
	if n != 2 {
		t.Fatalf("want 2 rows, got %d", n)
	}
}

func TestDeleteWaterLogByClientID(t *testing.T) {
	db := testDB(t)
	day := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	id := "6f1c2c1e-0000-4000-8000-000000000009"
	if _, err := CreateWaterLog(asUser(db, 1), day, 8, nil, id); err != nil {
		t.Fatal(err)
	}
	if err := DeleteWaterLogByClientID(asUser(db, 1), id); err != nil {
		t.Fatal(err)
	}
	if err := DeleteWaterLogByClientID(asUser(db, 1), id); err != database.ErrNotFound {
		t.Fatalf("second delete: want not found, got %v", err)
	}
}
