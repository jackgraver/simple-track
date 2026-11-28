// Package pgtest gives tests throwaway Postgres databases on TEST_DATABASE_URL.
//
// Without TEST_DATABASE_URL (or with it unreachable) the tests that need Postgres are skipped,
// unless TEST_DATABASE_REQUIRED=1, which turns the skip into a failure. The j-cicd backend-test
// step sets both (dev/ci/backend-test.sh), so Postgres-only tests can't silently stop
// running in CI. `./bin/dev test` sets TEST_DATABASE_URL to the slot's Postgres (owner role).
//
// Migrated databases are copied from a template that is built once per migration set (named by
// a hash of the migrations), so a test costs a CREATE DATABASE, not a full migrate.
//
// This package must not import internal/database: that package's own tests use it.
package pgtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"be-simpletracker/internal/migrate"
	"be-simpletracker/migrations"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// AppRole is the server's login role (internal/database.AppRole; duplicated to avoid the import).
const AppRole = "simpletrack_app"

// Required reports whether Postgres tests must run (TEST_DATABASE_REQUIRED=1): a missing or
// unreachable TEST_DATABASE_URL fails the test instead of skipping it.
func Required() bool {
	v := strings.ToLower(os.Getenv("TEST_DATABASE_REQUIRED"))
	return v == "1" || v == "true" || v == "yes"
}

func skipOrFail(t testing.TB, format string, args ...any) {
	t.Helper()
	if Required() {
		t.Fatalf("TEST_DATABASE_REQUIRED is set: "+format, args...)
	}
	t.Skipf(format, args...)
}

// Admin returns a connection to TEST_DATABASE_URL (a role that can create databases, normally
// the owner/superuser) and its parsed URL, closed at cleanup. It skips the test (or fails it, see
// Required) when Postgres isn't available.
func Admin(t testing.TB) (*sql.DB, *url.URL) {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		skipOrFail(t, "TEST_DATABASE_URL not set")
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("TEST_DATABASE_URL: %v", err)
	}
	db, err := sql.Open("pgx", raw)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Ping(); err != nil {
		skipOrFail(t, "TEST_DATABASE_URL unreachable: %v", err)
	}
	return db, u
}

// NewDatabase creates an empty database (dropped at cleanup) and returns its DSN, on the same
// server and role as TEST_DATABASE_URL. prefix names it (letters, digits, underscores).
func NewDatabase(t testing.TB, prefix string) string {
	t.Helper()
	return createDB(t, prefix, "")
}

// Migrated creates a database with every migration applied (dropped at cleanup) and returns its
// DSN. It is a copy of a per-migration-set template, so it's cheap.
func Migrated(t testing.TB, prefix string) string {
	t.Helper()
	tpl := template(t)
	return createDB(t, prefix, tpl)
}

func createDB(t testing.TB, prefix, tpl string) string {
	t.Helper()
	admin, u := Admin(t)
	name := fmt.Sprintf("%s_%d", prefix, rand.Uint32())
	stmt := "CREATE DATABASE " + pgx.Identifier{name}.Sanitize()
	if tpl == "" {
		if _, err := admin.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	} else {
		// Clone under the template lock (shared), so another test process can't rebuild or drop
		// the template (exclusive) mid-copy.
		stmt += " TEMPLATE " + pgx.Identifier{tpl}.Sanitize()
		ctx := context.Background()
		conn, err := admin.Conn(ctx)
		if err != nil {
			t.Fatal(err)
		}
		_, err = conn.ExecContext(ctx, "SELECT pg_advisory_lock_shared($1)", templateLockKey)
		if err == nil {
			_, err = conn.ExecContext(ctx, stmt)
			conn.ExecContext(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock_shared($1)", templateLockKey)
		}
		conn.Close()
		if err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	// Cleanups run last-in first-out: this runs after the pools opened on the database are closed
	// and before admin is.
	t.Cleanup(func() {
		admin.Exec("DROP DATABASE IF EXISTS " + pgx.Identifier{name}.Sanitize() + " WITH (FORCE)")
	})
	dsn := *u
	dsn.Path = "/" + name
	return dsn.String()
}

var (
	tplMu   sync.Mutex
	tplName string
)

// templateLockKey guards templates across test processes (go test runs packages in parallel):
// building or dropping one takes it exclusively, cloning one shares it. Any value distinct from
// internal/migrate's lock works.
const templateLockKey int64 = 0x70_67_74_70_6c // "pgtpl"

// template returns the name of a fully migrated template database for the current migration
// set, building it if needed. A template is only marked IS_TEMPLATE once migrated, so a build
// that died halfway is rebuilt. Templates for older migration sets are dropped.
func template(t testing.TB) string {
	t.Helper()
	tplMu.Lock()
	defer tplMu.Unlock()
	if tplName != "" {
		return tplName
	}
	admin, u := Admin(t)
	ms, err := migrate.Load(migrations.FS)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	for _, m := range ms {
		fmt.Fprintf(h, "%d:%s\n", m.Version, m.Checksum)
	}
	name := "pgtest_tpl_" + hex.EncodeToString(h.Sum(nil))[:12]

	ctx := context.Background()
	conn, err := admin.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", templateLockKey); err != nil {
		t.Fatal(err)
	}
	defer conn.ExecContext(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", templateLockKey)

	var ready bool
	err = conn.QueryRowContext(ctx, "SELECT datistemplate FROM pg_database WHERE datname = $1", name).Scan(&ready)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	if !ready {
		if err := buildTemplate(ctx, conn, u, name); err != nil {
			t.Fatalf("build template %s: %v", name, err)
		}
	}
	dropOldTemplates(ctx, conn, name)
	tplName = name
	return name
}

func buildTemplate(ctx context.Context, conn *sql.Conn, u *url.URL, name string) error {
	q := pgx.Identifier{name}.Sanitize()
	if _, err := conn.ExecContext(ctx, "DROP DATABASE IF EXISTS "+q+" WITH (FORCE)"); err != nil {
		return err
	}
	if _, err := conn.ExecContext(ctx, "CREATE DATABASE "+q); err != nil {
		return err
	}
	dsn := *u
	dsn.Path = "/" + name
	db, err := sql.Open("pgx", dsn.String())
	if err != nil {
		return err
	}
	m, err := migrate.New(db, migrations.FS, nil)
	if err == nil {
		_, err = m.Up(ctx, 0)
	}
	db.Close()
	if err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, "ALTER DATABASE "+q+" IS_TEMPLATE true")
	return err
}

func dropOldTemplates(ctx context.Context, conn *sql.Conn, keep string) {
	rows, err := conn.QueryContext(ctx, `SELECT datname FROM pg_database WHERE datname LIKE 'pgtest\_tpl\_%' AND datname <> $1`, keep)
	if err != nil {
		return
	}
	var old []string
	for rows.Next() {
		var n string
		if rows.Scan(&n) == nil {
			old = append(old, n)
		}
	}
	rows.Close()
	for _, n := range old {
		q := pgx.Identifier{n}.Sanitize()
		conn.ExecContext(ctx, "ALTER DATABASE "+q+" IS_TEMPLATE false")
		conn.ExecContext(ctx, "DROP DATABASE IF EXISTS "+q+" WITH (FORCE)")
	}
}

func gormConfig() *gorm.Config {
	return &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
}

// Open connects to dsn as its own role (the owner, for TEST_DATABASE_URL databases), closed at
// cleanup. The owner bypasses row-level security: use it to arrange and inspect, not to test
// what the app can see.
func Open(t testing.TB, dsn string) *gorm.DB {
	t.Helper()
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	return openGorm(t, sqlDB)
}

// OpenAsApp connects to dsn and runs every session as the server's app role (SET ROLE
// simpletrack_app on connect), so row-level security applies as it does to the real server.
// The role's login settings (statement_timeout, ...) don't apply after SET ROLE.
func OpenAsApp(t testing.TB, dsn string) *gorm.DB {
	t.Helper()
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB := stdlib.OpenDB(*cfg, stdlib.OptionAfterConnect(func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, "SET ROLE "+AppRole)
		return err
	}))
	return openGorm(t, sqlDB)
}

func openGorm(t testing.TB, sqlDB *sql.DB) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), gormConfig())
	if err != nil {
		sqlDB.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	return db
}
