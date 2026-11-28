package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"testing/fstest"

	"be-simpletracker/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// skipPG skips a test that needs Postgres, or fails it when TEST_DATABASE_REQUIRED is set (CI).
// The same rule as internal/database/pgtest, which this package can't import (pgtest imports it).
func skipPG(t *testing.T, format string, args ...any) {
	t.Helper()
	switch strings.ToLower(os.Getenv("TEST_DATABASE_REQUIRED")) {
	case "1", "true", "yes":
		t.Fatalf("TEST_DATABASE_REQUIRED is set: "+format, args...)
	}
	t.Skipf(format, args...)
}

// testDB creates a fresh, empty database for one test and drops it afterwards.
// It needs TEST_DATABASE_URL pointing at a Postgres server where the user may
// CREATE DATABASE (`dev test` sets it to the dev database); otherwise the
// test is skipped.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	admin := os.Getenv("TEST_DATABASE_URL")
	if admin == "" {
		skipPG(t, "TEST_DATABASE_URL not set")
	}
	adminDB, err := sql.Open("pgx", admin)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { adminDB.Close() })
	if err := adminDB.Ping(); err != nil {
		skipPG(t, "TEST_DATABASE_URL unreachable: %v", err)
	}

	name := fmt.Sprintf("migrate_test_%d", rand.Uint32())
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
	return db
}

func files(m map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for name, content := range m {
		fsys[name] = &fstest.MapFile{Data: []byte(content)}
	}
	return fsys
}

var twoTables = map[string]string{
	"0001_a.sql": "-- migrate:up\nCREATE TABLE a (id int);\nINSERT INTO a VALUES (1);\n-- migrate:down\nDROP TABLE a;\n",
	"0002_b.sql": "-- migrate:up\nCREATE TABLE b (id int);\n-- migrate:down\nDROP TABLE b;\n",
}

func newMigrator(t *testing.T, db *sql.DB, m map[string]string) *Migrator {
	t.Helper()
	mig, err := New(db, files(m), t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	return mig
}

func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()
	var ok bool
	if err := db.QueryRow(`SELECT to_regclass($1) IS NOT NULL`, name).Scan(&ok); err != nil {
		t.Fatal(err)
	}
	return ok
}

func TestUpDownCheck(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	m := newMigrator(t, db, twoTables)

	if err := m.Check(ctx); err == nil || !strings.Contains(err.Error(), "no schema_migrations") {
		t.Fatalf("check on empty db: %v", err)
	}

	done, err := m.Up(ctx, 1)
	if err != nil || len(done) != 1 {
		t.Fatalf("up to 1: %v %v", done, err)
	}
	if err := m.Check(ctx); err == nil || !strings.Contains(err.Error(), "1 pending") {
		t.Fatalf("check with pending: %v", err)
	}

	if _, err := m.Up(ctx, 0); err != nil {
		t.Fatal(err)
	}
	if !tableExists(t, db, "a") || !tableExists(t, db, "b") {
		t.Fatal("tables missing after up")
	}
	if err := m.Check(ctx); err != nil {
		t.Fatalf("check after up: %v", err)
	}
	if done, err := m.Up(ctx, 0); err != nil || len(done) != 0 {
		t.Fatalf("second up: %v %v", done, err)
	}

	if _, err := m.Down(ctx, 2); err != nil {
		t.Fatal(err)
	}
	if tableExists(t, db, "a") || tableExists(t, db, "b") {
		t.Fatal("tables remain after down")
	}
	if s, _ := m.Status(ctx); s.Current() != 0 || len(s.Pending) != 2 {
		t.Fatalf("status after down: %+v", s)
	}
	if _, err := m.Down(ctx, 1); err == nil {
		t.Fatal("down with nothing applied succeeded")
	}
}

// A failing migration must leave neither its changes nor a history row.
func TestFailedMigrationRollsBack(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	m := newMigrator(t, db, map[string]string{
		"0001_a.sql": twoTables["0001_a.sql"],
		"0002_bad.sql": "-- migrate:up\nCREATE TABLE half (id int);\nSELECT * FROM does_not_exist;\n" +
			"-- migrate:down\nDROP TABLE half;\n",
	})
	_, err := m.Up(ctx, 0)
	if err == nil || !strings.Contains(err.Error(), "0002_bad.sql (up)") {
		t.Fatalf("got %v", err)
	}
	if tableExists(t, db, "half") {
		t.Error("partial migration was committed")
	}
	if s, _ := m.Status(ctx); s.Current() != 1 {
		t.Errorf("current = %d, want 1", s.Current())
	}
}

func TestChangedAppliedMigrationIsRefused(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	if _, err := newMigrator(t, db, twoTables).Up(ctx, 0); err != nil {
		t.Fatal(err)
	}
	edited := map[string]string{
		"0001_a.sql": twoTables["0001_a.sql"],
		"0002_b.sql": "-- migrate:up\nCREATE TABLE b (id bigint);\n-- migrate:down\nDROP TABLE b;\n",
	}
	m := newMigrator(t, db, edited)
	if err := m.Check(ctx); err == nil || !strings.Contains(err.Error(), "0002_b.sql") {
		t.Fatalf("check: %v", err)
	}
	if _, err := m.Up(ctx, 0); err == nil {
		t.Fatal("up ran with an edited applied migration")
	}
	// The dev fix: roll the latest one back with its (current) down, re-apply.
	if _, err := m.Down(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(ctx, 0); err != nil {
		t.Fatal(err)
	}
	if err := m.Check(ctx); err != nil {
		t.Fatal(err)
	}
}

func TestIrreversibleBlocksWholeRollback(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	m := newMigrator(t, db, map[string]string{
		"0001_a.sql": "-- migrate:up\nCREATE TABLE a (id int);\n-- migrate:down irreversible\n",
		"0002_b.sql": twoTables["0002_b.sql"],
	})
	if _, err := m.Up(ctx, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := m.Down(ctx, 2); err == nil || !strings.Contains(err.Error(), "irreversible") {
		t.Fatalf("got %v", err)
	}
	if !tableExists(t, db, "b") {
		t.Error("0002 was rolled back even though the run was refused")
	}
}

func TestDatabaseAheadOfBinary(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	if _, err := newMigrator(t, db, twoTables).Up(ctx, 0); err != nil {
		t.Fatal(err)
	}
	old := newMigrator(t, db, map[string]string{"0001_a.sql": twoTables["0001_a.sql"]})
	if err := old.Check(ctx); err != nil {
		t.Fatalf("an older binary must still start: %v", err)
	}
	if s, _ := old.Status(ctx); len(s.Unknown) != 1 {
		t.Errorf("unknown = %+v", s.Unknown)
	}
	if _, err := old.Down(ctx, 1); err == nil || !strings.Contains(err.Error(), "no file") {
		t.Fatalf("down of unknown migration: %v", err)
	}
}

func TestOutOfOrderPendingIsRefused(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	if _, err := newMigrator(t, db, twoTables).Up(ctx, 0); err != nil {
		t.Fatal(err)
	}
	// Simulate 0001 never having been recorded while 0002 was.
	if _, err := db.Exec(`DELETE FROM schema_migrations WHERE version = 1`); err != nil {
		t.Fatal(err)
	}
	if _, err := newMigrator(t, db, twoTables).Up(ctx, 0); err == nil || !strings.Contains(err.Error(), "in order") {
		t.Fatalf("got %v", err)
	}
}

func TestUntrackedSchemaNeedsBaseline(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	if _, err := db.Exec(`CREATE TABLE a (id int)`); err != nil {
		t.Fatal(err)
	}
	m := newMigrator(t, db, twoTables)
	if _, err := m.Up(ctx, 0); err == nil || !strings.Contains(err.Error(), "baseline") {
		t.Fatalf("got %v", err)
	}
	if err := m.Baseline(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if err := m.Baseline(ctx, 1); err == nil {
		t.Fatal("second baseline succeeded")
	}
	if done, err := m.Up(ctx, 0); err != nil || len(done) != 1 || done[0].Version != 2 {
		t.Fatalf("up after baseline: %v %v", done, err)
	}
}

func TestNoTransactionMigration(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	m := newMigrator(t, db, map[string]string{
		"0001_a.sql": twoTables["0001_a.sql"],
		"0002_idx.sql": "-- migrate:no-transaction\n-- migrate:up\nCREATE INDEX CONCURRENTLY IF NOT EXISTS a_id ON a (id);\n" +
			"-- migrate:down\nDROP INDEX CONCURRENTLY IF EXISTS a_id;\n",
	})
	if _, err := m.Up(ctx, 0); err != nil {
		t.Fatal(err)
	}
	if !tableExists(t, db, "a_id") {
		t.Fatal("index missing")
	}
	if _, err := m.Down(ctx, 1); err != nil {
		t.Fatal(err)
	}
}

// Two migrators racing must apply each migration exactly once.
func TestConcurrentUpIsSerialised(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for range 4 {
		wg.Go(func() {
			_, err := newMigrator(t, db, twoTables).Up(ctx, 0)
			errs <- err
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Error(err)
		}
	}
	var rows int
	db.QueryRow(`SELECT count(*) FROM a`).Scan(&rows)
	if rows != 1 {
		t.Errorf("0001 ran %d times", rows)
	}
}

// The real baseline must apply cleanly to an empty database and leave the
// server's startup check passing.
func TestEmbeddedMigrationsApply(t *testing.T) {
	ctx := context.Background()
	db := testDB(t)
	m, err := New(db, migrations.FS, t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(ctx, 0); err != nil {
		t.Fatal(err)
	}
	if err := m.Check(ctx); err != nil {
		t.Fatal(err)
	}
	var active int
	db.QueryRow(`SELECT count(*) FROM workout_programs WHERE is_active`).Scan(&active)
	if active != 1 {
		t.Errorf("active workout programs = %d, want 1", active)
	}
}
