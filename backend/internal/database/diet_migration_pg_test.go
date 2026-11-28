package database_test

import (
	"strings"
	"testing"

	"be-simpletracker/internal/database/pgtest"
)

// C2's diet migration on legacy-shaped data: every row (soft-deleted too) goes to the admin, none is
// lost, children keep their parents, a dangling legacy composite_food_id doesn't stop the migration
// (its FK stays NOT VALID), and it rolls back cleanly.
func TestDietMigrationBackfill(t *testing.T) {
	before := migrationVersion(t, "diet_owned") - 1
	seed := `INSERT INTO users (username, password, email, is_admin, created_at) VALUES
			('owner', 'x', 'o@x', true, now()), ('other', 'x', 'p@x', false, now());
		INSERT INTO foods (name, serving_type, serving_amount, calories, quick_entry, created_at, deleted_at) VALUES
			('Rice', 'g', 100, 130, false, now(), NULL),
			('Snack [ql-1]', '', 1, 300, true, now(), NULL),
			('Gone', 'g', 1, 1, false, now(), now());
		INSERT INTO composite_foods (name, created_at) VALUES ('Pico', now());
		INSERT INTO composite_food_items (composite_food_id, food_id, amount, created_at) VALUES (1, 1, 2, now());
		INSERT INTO plans (name, calories, created_at) VALUES ('Plan', 2000, now());
		INSERT INTO days (date, plan_id, created_at) VALUES ('2026-01-01', 1, now()), ('2026-01-02', 1, now());
		INSERT INTO meals (name, client_id, created_at) VALUES
			('Lunch', '6f1c2c1e-0000-4000-8000-000000000001', now()), ('Old', NULL, now());
		UPDATE meals SET deleted_at = now() WHERE name = 'Old';
		INSERT INTO meal_items (meal_id, food_id, amount, composite_food_id, created_at) VALUES
			(1, 1, 1, 1, now()), (1, 2, 1, NULL, now()), (2, 1, 1, 999, now());
		INSERT INTO saved_meals (name, created_at) VALUES ('Bowl', now());
		INSERT INTO saved_meal_items (saved_meal_id, food_id, amount, created_at) VALUES (1, 1, 1, now());
		INSERT INTO planned_meals (day_id, meal_id, saved_meal_id, created_at) VALUES (1, 1, 1, now());
		INSERT INTO day_logs (day_id, meal_id, created_at) VALUES (2, 1, now());`
	tables := []string{"foods", "composite_foods", "composite_food_items", "plans", "days", "meals",
		"meal_items", "saved_meals", "saved_meal_items", "planned_meals", "day_logs"}

	dsn := pgtest.NewDatabase(t, "diet_backfill")
	if err := migrateTo(t, dsn, before); err != nil {
		t.Fatal(err)
	}
	db := pgtest.Open(t, dsn)
	if err := db.Exec(seed).Error; err != nil {
		t.Fatal(err)
	}
	count := func(table string) (n int64) {
		if err := db.Raw("SELECT count(*) FROM " + table).Scan(&n).Error; err != nil {
			t.Fatal(err)
		}
		return n
	}
	was := map[string]int64{}
	for _, tb := range tables {
		was[tb] = count(tb)
	}

	if err := migrateTo(t, dsn, 0); err != nil {
		t.Fatal(err)
	}
	for _, tb := range tables {
		if n := count(tb); n != was[tb] {
			t.Errorf("%s: %d rows before, %d after", tb, was[tb], n)
		}
		col := "user_id"
		if tb == "foods" {
			col = "created_by"
		}
		var wrong int64
		if err := db.Raw("SELECT count(*) FROM " + tb + " WHERE " + col + " <> (SELECT id FROM users WHERE username = 'owner')").
			Scan(&wrong).Error; err != nil || wrong != 0 {
			t.Errorf("%s: %d rows not given to the admin (%v)", tb, wrong, err)
		}
	}
	// The clean legacy links validated; the dangling composite_food_id left its FK NOT VALID.
	var invalid []string
	if err := db.Raw(`SELECT conname FROM pg_constraint WHERE NOT convalidated AND conrelid::regclass::text IN
		('planned_meals','meal_items','saved_meal_items') ORDER BY 1`).Scan(&invalid).Error; err != nil {
		t.Fatal(err)
	}
	if len(invalid) != 1 || invalid[0] != "fk_meal_items_composite_food" {
		t.Errorf("NOT VALID constraints %v, want only fk_meal_items_composite_food", invalid)
	}

	if err := migrateDown(t, dsn, 1); err != nil {
		t.Errorf("down: %v", err)
	}

	// With rows to give away and no admin, the migration stops rather than guess.
	noAdmin := pgtest.NewDatabase(t, "diet_noadmin")
	if err := migrateTo(t, noAdmin, before); err != nil {
		t.Fatal(err)
	}
	if err := pgtest.Open(t, noAdmin).Exec(strings.Replace(seed, "true", "false", 1)).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateTo(t, noAdmin, 0); err == nil || !strings.Contains(err.Error(), "expected exactly one") {
		t.Fatalf("migration with rows and no admin: %v, want it to stop", err)
	}
}
