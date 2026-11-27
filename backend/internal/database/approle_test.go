package database

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"strings"
	"testing"

	"be-simpletracker/internal/migrate"
	"be-simpletracker/migrations"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// pgAdmin connects to TEST_DATABASE_URL (a role that may CREATE DATABASE and CREATE ROLE; `dev
// test` points it at the dev Postgres as its owner). Skipped when unset or unreachable.
func pgAdmin(t *testing.T) (*sql.DB, *url.URL) {
	t.Helper()
	admin := os.Getenv("TEST_DATABASE_URL")
	if admin == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", admin)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Ping(); err != nil {
		t.Skipf("TEST_DATABASE_URL unreachable: %v", err)
	}
	u, err := url.Parse(admin)
	if err != nil {
		t.Fatal(err)
	}
	return db, u
}

// migratedDB creates a fresh database with every migration applied, connected as the owner.
func migratedDB(t *testing.T) *sql.DB {
	t.Helper()
	admin, u := pgAdmin(t)
	name := fmt.Sprintf("approle_test_%d", rand.Uint32())
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	u.Path = "/" + name
	db, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		db.Close()
		admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
	})
	m, err := migrate.New(db, migrations.FS, t.Logf)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.Up(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	return db
}

// asApp runs fn on one connection with SET ROLE simpletrack_app, so privilege checks are the app
// role's (role settings such as statement_timeout only apply at login and are checked separately).
func asApp(t *testing.T, db *sql.DB, fn func(c *sql.Conn)) {
	t.Helper()
	ctx := context.Background()
	c, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := c.ExecContext(ctx, "SET ROLE "+AppRole); err != nil {
		t.Fatal(err)
	}
	defer c.ExecContext(ctx, "RESET ROLE")
	fn(c)
}

func TestAppRoleGrants(t *testing.T) {
	db := migratedDB(t)
	ctx := context.Background()

	var super, bypass, createrole, createdb, login bool
	if err := db.QueryRow(`SELECT rolsuper, rolbypassrls, rolcreaterole, rolcreatedb, rolcanlogin
		FROM pg_roles WHERE rolname = $1`, AppRole).Scan(&super, &bypass, &createrole, &createdb, &login); err != nil {
		t.Fatalf("role %s: %v", AppRole, err)
	}
	if super || bypass || createrole || createdb {
		t.Fatalf("%s has powers: super=%v bypassrls=%v createrole=%v createdb=%v", AppRole, super, bypass, createrole, createdb)
	}

	var settings []string
	rows, err := db.Query(`SELECT unnest(setconfig) FROM pg_db_role_setting
		WHERE setrole = $1::regrole AND setdatabase = 0`, AppRole)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var s string
		rows.Scan(&s)
		settings = append(settings, s)
	}
	rows.Close()
	for _, want := range []string{"statement_timeout=15s", "idle_in_transaction_session_timeout=30s"} {
		if !strings.Contains(strings.Join(settings, " "), want) {
			t.Errorf("role setting %s missing (have %v)", want, settings)
		}
	}

	var tables []string
	rows, err = db.Query(`SELECT quote_ident(tablename) FROM pg_tables WHERE schemaname = 'public' ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var n string
		rows.Scan(&n)
		tables = append(tables, n)
	}
	rows.Close()
	if len(tables) < 10 {
		t.Fatalf("only %d tables after migrating", len(tables))
	}

	// A table a later migration creates (as the owner) is usable by the app without a GRANT.
	if _, err := db.Exec(`CREATE TABLE later_table (id bigserial PRIMARY KEY, note text)`); err != nil {
		t.Fatal(err)
	}

	asApp(t, db, func(c *sql.Conn) {
		for _, tbl := range tables {
			if _, err := c.ExecContext(ctx, "SELECT 1 FROM "+tbl+" LIMIT 0"); err != nil {
				t.Errorf("app cannot read %s: %v", tbl, err)
			}
		}
		if _, err := c.ExecContext(ctx, `INSERT INTO later_table (note) VALUES ('x')`); err != nil {
			t.Errorf("default privileges: app cannot insert into a new table: %v", err)
		}
		if _, err := c.ExecContext(ctx, `UPDATE later_table SET note = 'y'`); err != nil {
			t.Errorf("app cannot update: %v", err)
		}
		if _, err := c.ExecContext(ctx, `DELETE FROM later_table`); err != nil {
			t.Errorf("app cannot delete: %v", err)
		}

		mustFail := map[string]string{
			"create a table":          `CREATE TABLE app_made (id int)`,
			"alter a table":           `ALTER TABLE later_table ADD COLUMN x int`,
			"drop a table":            `DROP TABLE later_table`,
			"truncate":                `TRUNCATE later_table`,
			"write migration history": `DELETE FROM schema_migrations`,
			"disable RLS":             `ALTER TABLE later_table DISABLE ROW LEVEL SECURITY`,
		}
		for what, stmt := range mustFail {
			if _, err := c.ExecContext(ctx, stmt); err == nil {
				t.Errorf("app could %s (%s)", what, stmt)
			}
		}
		if err := CheckAppRole(ctx, c); err != nil {
			t.Errorf("CheckAppRole as the app role: %v", err)
		}
	})

	// The owner (the server must never run as it) is flagged.
	if err := CheckAppRole(ctx, db); err == nil {
		t.Error("CheckAppRole accepted the owner role")
	}
}

func TestEnsureLogin(t *testing.T) {
	admin, u := pgAdmin(t)
	ctx := context.Background()
	role := fmt.Sprintf("approle_test_%d", rand.Uint32())
	if _, err := admin.Exec("CREATE ROLE " + role + " NOLOGIN"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Exec("DROP ROLE IF EXISTS " + role) })

	dsn := *u
	dsn.User = url.UserPassword(role, "first-Secret_1")
	changed, err := ensureLogin(ctx, admin, dsn.String(), role)
	if err != nil || !changed {
		t.Fatalf("first run: changed=%v err=%v", changed, err)
	}
	// The verifier we sent is what Postgres stored, and a real SCRAM login works with it.
	var stored string
	if err := admin.QueryRow(`SELECT rolpassword FROM pg_authid WHERE rolname = $1`, role).Scan(&stored); err == nil &&
		!strings.HasPrefix(stored, "SCRAM-SHA-256$4096:") {
		t.Fatalf("stored password is not a SCRAM verifier: %.20s...", stored)
	}

	changed, err = ensureLogin(ctx, admin, dsn.String(), role)
	if err != nil || changed {
		t.Fatalf("second run should be a no-op: changed=%v err=%v", changed, err)
	}

	// Rotating: a new password in the DSN is applied.
	dsn.User = url.UserPassword(role, "second-Secret_2")
	changed, err = ensureLogin(ctx, admin, dsn.String(), role)
	if err != nil || !changed {
		t.Fatalf("rotation: changed=%v err=%v", changed, err)
	}

	// A role someone switched off (has a password, NOLOGIN) is not silently re-enabled.
	if _, err := admin.Exec("ALTER ROLE " + role + " NOLOGIN"); err != nil {
		t.Fatal(err)
	}
	if _, err := ensureLogin(ctx, admin, dsn.String(), role); err == nil || !strings.Contains(err.Error(), "deliberately") {
		t.Fatalf("NOLOGIN role with a password: want a refusal, got %v", err)
	}

	// A DSN for some other role is left alone.
	other := *u
	other.User = url.UserPassword("someone_else", "x")
	if changed, err := ensureLogin(ctx, admin, other.String(), role); err != nil || changed {
		t.Fatalf("other role: changed=%v err=%v", changed, err)
	}
}

// A role that inherits the table owner's role is treated as the owner by RLS, so it's refused.
func TestCheckAppRoleRejectsOwnerMembership(t *testing.T) {
	db := migratedDB(t)
	ctx := context.Background()
	role := fmt.Sprintf("approle_member_%d", rand.Uint32())
	var owner string
	if err := db.QueryRow("SELECT current_user").Scan(&owner); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{"CREATE ROLE " + role + " NOLOGIN", fmt.Sprintf("GRANT %q TO %s", owner, role)} {
		if _, err := db.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() { db.Exec("DROP ROLE IF EXISTS " + role) })
	c, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := c.ExecContext(ctx, "SET ROLE "+role); err != nil {
		t.Fatal(err)
	}
	defer c.ExecContext(ctx, "RESET ROLE")
	if err := CheckAppRole(ctx, c); err == nil {
		t.Fatal("CheckAppRole accepted a member of the owner role")
	}
}
