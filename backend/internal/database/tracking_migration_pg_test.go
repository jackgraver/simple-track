package database_test

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"be-simpletracker/internal/database/pgtest"
	"be-simpletracker/internal/migrate"
	"be-simpletracker/migrations"
)

func migrationVersion(t *testing.T, name string) int {
	t.Helper()
	ms, err := migrate.Load(migrations.FS)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range ms {
		if m.Name == name {
			return m.Version
		}
	}
	t.Fatalf("no %s migration", name)
	return 0
}

// C1's tracking migration on legacy-shaped data: every row (soft-deleted too) goes to the admin and
// none is lost, a water log keeps its preset, and a user_profiles table that somehow holds more
// than one live row ends with one live row per user (the oldest, which is the one the app read)
// and the others soft-deleted, not dropped.
func TestTrackingMigrationBackfill(t *testing.T) {
	before := migrationVersion(t, "tracking_owned") - 1
	seed := `INSERT INTO users (username, password, email, is_admin, created_at) VALUES
			('owner', 'x', 'o@x', true, now()), ('other', 'x', 'p@x', false, now());
		INSERT INTO body_weight_logs (date, weight_lbs, created_at, deleted_at) VALUES
			('2026-01-01', 180, now(), NULL), ('2026-01-02', 181, now(), now());
		INSERT INTO step_logs (date, steps, created_at) VALUES ('2026-01-01', 5000, now());
		INSERT INTO drink_size_presets (name, amount_oz, created_at, user_id)
			SELECT 'Glass', 8, now(), id FROM users WHERE username = 'owner';
		INSERT INTO water_logs (date, amount_oz, preset_id, client_id, created_at) VALUES
			('2026-01-01', 8, 1, '6f1c2c1e-0000-4000-8000-000000000001', now()),
			('2026-01-01', 8, NULL, NULL, now());
		INSERT INTO grocery_items (name, client_id, created_at, deleted_at) VALUES
			('Milk', '11111111-1111-4111-8111-111111111111', now(), NULL), ('Old', NULL, now(), now());
		INSERT INTO user_profiles (height_in, age, sex, activity_level, created_at, deleted_at) VALUES
			(70, 30, 'male', 'sedentary', now(), NULL),
			(71, 31, 'male', 'sedentary', now(), NULL),
			(72, 32, 'male', 'sedentary', now(), now());`
	tables := []string{"body_weight_logs", "step_logs", "water_logs", "grocery_items", "user_profiles"}

	dsn := pgtest.NewDatabase(t, "tracking_backfill")
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
		var wrong int64
		if err := db.Raw("SELECT count(*) FROM " + tb + " WHERE user_id <> (SELECT id FROM users WHERE username = 'owner')").
			Scan(&wrong).Error; err != nil || wrong != 0 {
			t.Errorf("%s: %d rows not given to the admin (%v)", tb, wrong, err)
		}
	}
	var live, withPreset int64
	if err := db.Raw("SELECT count(*) FROM user_profiles WHERE deleted_at IS NULL").Scan(&live).Error; err != nil || live != 1 {
		t.Errorf("%d live profiles, want 1 (%v)", live, err)
	}
	var height float64
	if err := db.Raw("SELECT height_in FROM user_profiles WHERE deleted_at IS NULL").Scan(&height).Error; err != nil || height != 70 {
		t.Errorf("surviving profile height %v, want the oldest (70) (%v)", height, err)
	}
	if err := db.Raw("SELECT count(*) FROM water_logs WHERE preset_id = 1").Scan(&withPreset).Error; err != nil || withPreset != 1 {
		t.Errorf("water log lost its preset (%d, %v)", withPreset, err)
	}

	// Rolls back cleanly.
	m := func() error { return migrateDown(t, dsn, 1) }
	if err := m(); err != nil {
		t.Errorf("down: %v", err)
	}

	// With rows to give away and no admin, the migration stops rather than guess.
	noAdmin := pgtest.NewDatabase(t, "tracking_noadmin")
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

// migrateDown rolls back the latest n migrations on dsn as the owner.
func migrateDown(t *testing.T, dsn string, n int) error {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	m, err := migrate.New(db, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = m.Down(context.Background(), n)
	return err
}
