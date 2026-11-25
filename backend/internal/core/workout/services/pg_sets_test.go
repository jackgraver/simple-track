package services_test

import (
	"be-simpletracker/internal/clock"
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/core/workout/services"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/migrate"
	"be-simpletracker/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// pgTestDB creates a fresh database on TEST_DATABASE_URL (skipped when unset), returning its
// URL and a *sql.DB. Same convention as internal/migrate's tests.
func pgTestDB(t *testing.T) (string, *sql.DB) {
	t.Helper()
	admin := os.Getenv("TEST_DATABASE_URL")
	if admin == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	adminDB, err := sql.Open("pgx", admin)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { adminDB.Close() })
	if err := adminDB.Ping(); err != nil {
		t.Skipf("TEST_DATABASE_URL unreachable: %v", err)
	}
	name := fmt.Sprintf("workout_sets_test_%d", rand.Uint32())
	if _, err := adminDB.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(admin)
	if err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	db, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		adminDB.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
	})
	return u.String(), db
}

func mustExec(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

func count(t *testing.T, db *sql.DB, q string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return n
}

// Migration 0004 merges duplicate logged_exercises, is reversible down to its indexes, and the
// set endpoints hold up under real concurrency on Postgres.
func TestPostgres_setsMigrationAndConcurrency(t *testing.T) {
	dsn, sqlDB := pgTestDB(t)
	ctx := context.Background()
	m, err := migrate.New(sqlDB, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(ctx, 3); err != nil {
		t.Fatal(err)
	}

	// The pre-0004 duplicate from the audit: two parents for one (log, exercise).
	mustExec(t, sqlDB, `INSERT INTO exercises (id, name) VALUES (1, 'Squat'), (2, 'Bench')`)
	mustExec(t, sqlDB, `INSERT INTO workout_logs (id, date) VALUES (1, '2026-01-05')`)
	mustExec(t, sqlDB, `INSERT INTO logged_exercises (id, workout_log_id, exercise_id, notes) VALUES
		(10, 1, 1, ''), (11, 1, 1, 'felt heavy'), (12, 1, 1, 'belt'), (13, 1, 2, 'solo')`)
	mustExec(t, sqlDB, `INSERT INTO logged_sets (logged_exercise_id, reps, weight) VALUES
		(10, 5, 100), (10, 5, 100), (11, 5, 100), (12, 3, 110), (13, 8, 60)`)
	for _, tbl := range []string{"exercises", "workout_logs", "logged_exercises"} {
		mustExec(t, sqlDB, fmt.Sprintf(`SELECT setval('%[1]s_id_seq', (SELECT max(id) FROM %[1]s))`, tbl))
	}

	if _, err := m.Up(ctx, 0); err != nil {
		t.Fatal(err)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM logged_exercises WHERE exercise_id = 1`); n != 1 {
		t.Fatalf("duplicates not merged: %d parents", n)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM logged_sets WHERE logged_exercise_id = 10`); n != 4 {
		t.Fatalf("sets not moved to the kept parent: %d", n)
	}
	var notes string
	if err := sqlDB.QueryRow(`SELECT notes FROM logged_exercises WHERE id = 10`).Scan(&notes); err != nil || notes != "felt heavy\nbelt" {
		t.Fatalf("notes = %q, %v", notes, err)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM logged_exercises WHERE id = 13 AND notes = 'solo'`); n != 1 {
		t.Fatalf("non-duplicate parent changed")
	}
	if _, err := sqlDB.Exec(`INSERT INTO logged_exercises (workout_log_id, exercise_id) VALUES (1, 1)`); err == nil {
		t.Fatalf("unique (workout_log_id, exercise_id) not enforced")
	}
	if _, err := m.Down(ctx, 1); err != nil {
		t.Fatalf("down: %v", err)
	}
	if _, err := m.Up(ctx, 0); err != nil {
		t.Fatalf("re-up: %v", err)
	}
	if err := m.Check(ctx); err != nil {
		t.Fatalf("check: %v", err)
	}

	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if d, err := gdb.DB(); err == nil {
		d.SetMaxOpenConns(40) // below Postgres' default max_connections (100)
	}
	database.SetDB(gdb)
	t.Cleanup(func() {
		if d, err := gdb.DB(); err == nil {
			d.Close()
		}
	})

	// Get-or-create of a day's workout log is serialized per date (no unique index on date).
	var dayWG sync.WaitGroup
	gate := make(chan struct{})
	dayErrs := make(chan error, 300)
	for d := 10; d < 30; d++ {
		for g := 0; g < 15; g++ {
			dayWG.Add(1)
			go func(d int) {
				defer dayWG.Done()
				<-gate
				_, err := services.GetOrCreateDay(ctx, clock.ZerodTime(d))
				dayErrs <- err
			}(d)
		}
	}
	close(gate)
	dayWG.Wait()
	close(dayErrs)
	for err := range dayErrs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM (SELECT date FROM workout_logs GROUP BY date HAVING count(*) > 1) d`); n != 0 {
		t.Fatalf("%d days got duplicate workout logs", n)
	}

	// Fresh day (no workout log yet): 24 parallel requests, 8 distinct sets each sent 3 times.
	day := clock.ZerodTime(0)
	var wg sync.WaitGroup
	errs := make(chan error, 24)
	firsts := make(chan bool, 24)
	start := make(chan struct{})
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			res, err := services.LogSet(ctx, services.LogSetInput{
				ClientID: clientID(500 + i%8), Day: day, ExerciseID: 2, Reps: 5, Weight: 80,
			})
			errs <- err
			firsts <- res.FirstSetOfDay
		}(i)
	}
	close(start)
	wg.Wait()
	close(errs)
	close(firsts)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	nFirst := 0
	for f := range firsts {
		if f {
			nFirst++
		}
	}
	// At least one create must report first_set_of_day (the client invalidates the activity
	// heatmap on it). Racing creates may both report it, which only costs an extra refetch.
	if nFirst < 1 {
		t.Fatalf("no create reported first_set_of_day")
	}
	t.Logf("first_set_of_day reported by %d of 8 creates", nFirst)
	if n := count(t, sqlDB, `SELECT count(*) FROM workout_logs WHERE date = $1`, day); n != 1 {
		t.Fatalf("want 1 workout log for the day, got %d", n)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM logged_exercises le JOIN workout_logs wl ON wl.id = le.workout_log_id
		WHERE wl.date = $1 AND le.exercise_id = 2`, day); n != 1 {
		t.Fatalf("want 1 parent, got %d", n)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM logged_sets WHERE client_id IS NOT NULL`); n != 8 {
		t.Fatalf("want 8 sets, got %d", n)
	}

	res, err := services.LogSet(ctx, services.LogSetInput{ClientID: clientID(500), Day: day, ExerciseID: 2})
	if err != nil || res.Created || len(res.Exercise.Sets) != 8 {
		t.Fatalf("replay after the burst: %+v, %v", res, err)
	}
	var loaded []models.LoggedSet
	if err := gdb.Where("client_id = ?", clientID(500)).Find(&loaded).Error; err != nil || len(loaded) != 1 {
		t.Fatalf("lookup by client_id on a uuid column: %v, %d", err, len(loaded))
	}
}

// pgMigratedGorm returns a gorm DB on a fresh, fully migrated database, installed as the
// package-wide connection.
func pgMigratedGorm(t *testing.T) (*gorm.DB, *sql.DB) {
	t.Helper()
	dsn, sqlDB := pgTestDB(t)
	m, err := migrate.New(sqlDB, migrations.FS, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	database.SetDB(gdb)
	t.Cleanup(func() {
		if d, err := gdb.DB(); err == nil {
			d.Close()
		}
	})
	return gdb, sqlDB
}

// waitForLockWait blocks until some backend is waiting on a row lock, so a test can order two
// transactions deterministically.
func waitForLockWait(t *testing.T, db *sql.DB) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if count(t, db, `SELECT count(*) FROM pg_stat_activity WHERE wait_event_type = 'Lock' AND datname = current_database()`) > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("no backend started waiting on a lock")
}

// Review finding 1: deleting an exercise's only set while a new set for the same exercise is
// being created must not cascade the new (committed) set away.
func TestPostgres_deleteLastSetDoesNotEatConcurrentCreate(t *testing.T) {
	_, sqlDB := pgMigratedGorm(t)
	ctx := context.Background()
	mustExec(t, sqlDB, `INSERT INTO exercises (id, name) VALUES (1, 'Squat')`)
	day := clock.ZerodTime(0)

	first, err := services.LogSet(ctx, services.LogSetInput{ClientID: clientID(1), Day: day, ExerciseID: 1, Reps: 5})
	if err != nil {
		t.Fatal(err)
	}
	parentID := first.Exercise.ID

	// B: a create that has inserted its set under the existing parent but not committed yet
	// (what CreateSet looks like between the set insert and commit).
	b, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.Exec(`SELECT id FROM logged_exercises WHERE id = $1 FOR KEY SHARE`, parentID); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Exec(`INSERT INTO logged_sets (logged_exercise_id, client_id, reps, weight) VALUES ($1, $2, 3, 0)`,
		parentID, clientID(2)); err != nil {
		t.Fatal(err)
	}

	// A: delete the only committed set. It must wait for B, then see B's set and keep the parent.
	done := make(chan workoutrepo.DeleteSetResult, 1)
	errc := make(chan error, 1)
	go func() {
		res, err := workoutrepo.DeleteSet(ctx, workoutrepo.SetRef{ClientID: clientID(1)})
		done <- res
		errc <- err
	}()
	waitForLockWait(t, sqlDB)
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
	res, err := <-done, <-errc
	if err != nil {
		t.Fatal(err)
	}
	if !res.Deleted || res.ExerciseRemoved {
		t.Fatalf("delete result %+v: want deleted, parent kept", res)
	}
	if n := count(t, sqlDB, `SELECT count(*) FROM logged_sets WHERE client_id = $1`, clientID(2)); n != 1 {
		t.Fatalf("concurrently created set was cascaded away (%d rows)", n)
	}

	// Stress the real code paths against each other: each round deletes the only set while a
	// new one is logged. Every successful create must still exist afterwards.
	for i := 0; i < 150; i++ {
		ex := uint(100 + i)
		mustExec(t, sqlDB, `INSERT INTO exercises (id, name) VALUES ($1, $2)`, ex, fmt.Sprintf("Ex %d", ex))
		old, newID := clientID(1000+2*i), clientID(1001+2*i)
		if _, err := services.LogSet(ctx, services.LogSetInput{ClientID: old, Day: day, ExerciseID: ex, Reps: 1}); err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		var delErr, logErr error
		start := make(chan struct{})
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_, delErr = workoutrepo.DeleteSet(ctx, workoutrepo.SetRef{ClientID: old})
		}()
		go func() {
			defer wg.Done()
			<-start
			_, logErr = services.LogSet(ctx, services.LogSetInput{ClientID: newID, Day: day, ExerciseID: ex, Reps: 2})
		}()
		close(start)
		wg.Wait()
		if delErr != nil || logErr != nil {
			t.Fatalf("round %d: delete %v, log %v", i, delErr, logErr)
		}
		if n := count(t, sqlDB, `SELECT count(*) FROM logged_sets WHERE client_id = $1`, newID); n != 1 {
			t.Fatalf("round %d: confirmed set lost", i)
		}
	}
}
