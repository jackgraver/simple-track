package services_test

import (
	"context"
	"strings"
	"testing"

	"be-simpletracker/internal/migrate"
	"be-simpletracker/migrations"
)

// Migration 0017 on legacy-shaped data: every row goes to the one admin, machine settings move to
// exercise_settings, the empty duplicate-date log is soft-deleted (the one prod had), and a
// duplicate that holds data stops the migration instead of being guessed at. Postgres only.
func TestPostgres_migration0017Backfill(t *testing.T) {
	_, sqlDB := pgTestDB(t)
	ctx := context.Background()
	m, err := migrate.New(sqlDB, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(ctx, 15); err != nil {
		t.Fatal(err)
	}
	mustExec(t, sqlDB, `INSERT INTO users (username, password, email, is_admin, token_version) VALUES ('owner', 'x', 'o@example.test', true, 1)`)
	mustExec(t, sqlDB, `INSERT INTO exercises (id, name, machine_settings) VALUES
		(1, 'Squat', '[{"label":"Seat","value":"4"}]'), (2, 'Bench', NULL)`)
	mustExec(t, sqlDB, `INSERT INTO workout_programs (id, name, is_active) VALUES (5, 'PPL', true)`)
	mustExec(t, sqlDB, `INSERT INTO workout_plans (id, name, workout_program_id) VALUES (7, 'Push', 5)`)
	mustExec(t, sqlDB, `INSERT INTO workout_plan_days (workout_plan_id, day_of_week) VALUES (7, 1)`)
	mustExec(t, sqlDB, `INSERT INTO workout_plan_exercises (workout_plan_id, exercise_id, display_order) VALUES (7, 1, 0), (7, 2, 1)`)
	// 124 and 125 are the empty duplicate pair; 126 is a normal day with data.
	mustExec(t, sqlDB, `INSERT INTO workout_logs (id, date, workout_plan_id) VALUES
		(124, '2026-04-03 06:00:00+00', 7), (125, '2026-04-03 06:00:00+00', 7), (126, '2026-04-04 06:00:00+00', 7)`)
	mustExec(t, sqlDB, `INSERT INTO logged_exercises (id, workout_log_id, exercise_id) VALUES (30, 126, 1)`)
	mustExec(t, sqlDB, `INSERT INTO logged_sets (logged_exercise_id, reps, weight) VALUES (30, 5, 100)`)
	mustExec(t, sqlDB, `INSERT INTO cardios (workout_log_id, minutes, type) VALUES (126, 20, 'Run')`)

	if _, err := m.Up(ctx, 0); err != nil {
		t.Fatalf("up: %v", err)
	}
	if err := m.Check(ctx); err != nil {
		t.Fatal(err)
	}
	owner := count(t, sqlDB, `SELECT id FROM users WHERE username = 'owner'`)
	for tbl, want := range map[string]int{
		"workout_programs": 2 /* PPL + the baseline Default */, "workout_plans": 1, "workout_plan_days": 1, "workout_plan_exercises": 2,
		"workout_logs": 3, "logged_exercises": 1, "logged_sets": 1, "cardios": 1,
	} {
		if n := count(t, sqlDB, `SELECT count(*) FROM `+tbl); n != want {
			t.Errorf("%s: %d rows after migrating, want %d", tbl, n, want)
		}
		if n := count(t, sqlDB, `SELECT count(*) FROM `+tbl+` WHERE user_id <> $1`, owner); n != 0 {
			t.Errorf("%s: %d rows not given to the admin", tbl, n)
		}
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM exercises WHERE created_by = $1`, owner); n != 2 {
		t.Errorf("%d exercises credited to the admin, want 2", n)
	}
	// Machine settings moved to the admin's per-user rows; exercises without any get none.
	if n := count(t, sqlDB, `SELECT count(*) FROM exercise_settings WHERE user_id = $1 AND exercise_id = 1
		AND machine_settings @> '[{"label":"Seat","value":"4"}]'`, owner); n != 1 {
		t.Errorf("machine settings not copied: %d", n)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM exercise_settings`); n != 1 {
		t.Errorf("%d settings rows, want 1", n)
	}
	// The empty extra log is soft-deleted, the oldest kept; the unique index holds from now on.
	if n := count(t, sqlDB, `SELECT count(*) FROM workout_logs WHERE deleted_at IS NULL AND date = '2026-04-03 06:00:00+00'`); n != 1 {
		t.Errorf("%d live logs on the duplicated date, want 1", n)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM workout_logs WHERE id = 125 AND deleted_at IS NOT NULL`); n != 1 {
		t.Errorf("duplicate log 125 not soft-deleted")
	}
	if _, err := sqlDB.Exec(`INSERT INTO workout_logs (date, user_id) VALUES ('2026-04-03 06:00:00+00', $1)`, owner); err == nil {
		t.Errorf("(user_id, date) unique not enforced")
	}
	// And it goes back down and up again.
	if _, err := m.Down(ctx, 1); err != nil {
		t.Fatalf("down: %v", err)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM exercises WHERE machine_settings IS NOT NULL`); n != 1 {
		t.Errorf("machine settings not restored on down: %d", n)
	}
	if _, err := m.Up(ctx, 0); err != nil {
		t.Fatalf("re-up: %v", err)
	}
}

func TestPostgres_migration0017StopsOnDuplicateWithData(t *testing.T) {
	_, sqlDB := pgTestDB(t)
	ctx := context.Background()
	m, err := migrate.New(sqlDB, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(ctx, 15); err != nil {
		t.Fatal(err)
	}
	mustExec(t, sqlDB, `INSERT INTO users (username, password, email, is_admin, token_version) VALUES ('owner', 'x', 'o@example.test', true, 1)`)
	mustExec(t, sqlDB, `INSERT INTO exercises (id, name) VALUES (1, 'Squat')`)
	mustExec(t, sqlDB, `INSERT INTO workout_logs (id, date) VALUES (1, '2026-04-03 06:00:00+00'), (2, '2026-04-03 06:00:00+00')`)
	mustExec(t, sqlDB, `INSERT INTO logged_exercises (workout_log_id, exercise_id) VALUES (2, 1)`)
	_, err = m.Up(ctx, 0)
	if err == nil || !strings.Contains(err.Error(), "merge them by hand") {
		t.Fatalf("migration over a duplicate with data: %v, want it to stop", err)
	}
}

// A fresh database has no admin and no workout data: the baseline's placeholder program goes.
func TestPostgres_migration0017FreshDatabase(t *testing.T) {
	_, sqlDB := pgTestDB(t)
	m, err := migrate.New(sqlDB, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM workout_programs`); n != 0 {
		t.Fatalf("%d programs on a fresh database, want 0", n)
	}
}
