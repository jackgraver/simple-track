package database_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/pgtest"
	"be-simpletracker/internal/database/twousers"
	"be-simpletracker/internal/migrate"
	"be-simpletracker/migrations"
)

// rlsExempt never get row-level security (README "Row-level security"): the migration history,
// and users, which the session lookup reads before the request has an identity.
var rlsExempt = map[string]string{
	"schema_migrations": "migration history, owner-only writes",
	"users":             "read by the session lookup before there is an identity",
}

// nullableOwnerColumn lists identity columns that may be NULL. C6's auth tables aren't owned
// data: an event can have no account (unknown username) and a link's creator can be deleted.
var nullableOwnerColumn = map[string]bool{
	"auth_events.user_id":          true,
	"user_setup_tokens.created_by": true,
}

// rlsPending lists tables not yet converted, with the lane that converts them. Each lane deletes
// its entries in the PR that adds the policies; D1 requires this to be empty. An entry whose
// table already has RLS (or no longer exists) only logs a reminder, so lanes can merge in any order.
var rlsPending = map[string]string{
	"foods": "C2", "composite_foods": "C2", "composite_food_items": "C2", "days": "C2",
	"day_logs": "C2", "plans": "C2", "meals": "C2", "meal_items": "C2", "planned_meals": "C2",
	"saved_meals": "C2", "saved_meal_items": "C2",

	"exercises": "C3", "workout_programs": "C3", "workout_plans": "C3", "workout_plan_days": "C3",
	"workout_plan_exercises": "C3", "workout_logs": "C3", "logged_exercises": "C3",
	"logged_sets": "C3", "cardios": "C3",
}

// Every table has row-level security with at least one policy, and every user_id / created_by
// column is NOT NULL with a foreign key to users.
func TestRLSCoverage(t *testing.T) {
	db := pgtest.Open(t, pgtest.Migrated(t, "rls_coverage"))
	var tables []struct {
		Name     string
		RLS      bool
		Policies int
	}
	err := db.Raw(`SELECT c.relname AS name, c.relrowsecurity AS rls,
			(SELECT count(*) FROM pg_policy p WHERE p.polrelid = c.oid) AS policies
		FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p') ORDER BY 1`).Scan(&tables).Error
	if err != nil {
		t.Fatal(err)
	}
	if len(tables) < 20 {
		t.Fatalf("only %d tables found; is the database migrated?", len(tables))
	}
	seen := map[string]bool{}
	for _, tb := range tables {
		seen[tb.Name] = true
		_, exempt := rlsExempt[tb.Name]
		lane, pending := rlsPending[tb.Name]
		switch {
		case exempt:
		case tb.RLS && tb.Policies == 0:
			t.Errorf("%s: row-level security is on but it has no policy (every query sees nothing)", tb.Name)
		case tb.RLS && pending:
			t.Logf("%s has row-level security now: delete it from rlsPending (%s)", tb.Name, lane)
		case !tb.RLS && !pending:
			t.Errorf("%s has no row-level security: add policies (DECISIONS.md, B3 owner kit) or, only with a reason, an rlsPending/rlsExempt entry", tb.Name)
		}
	}
	for name := range rlsPending {
		if !seen[name] {
			t.Logf("rlsPending has %s, which no longer exists: delete the entry", name)
		}
	}

	var cols []struct{ Table, Column, Nullable string }
	err = db.Raw(`SELECT table_name AS "table", column_name AS "column", is_nullable AS nullable
		FROM information_schema.columns
		WHERE table_schema = 'public' AND column_name IN ('user_id', 'created_by')`).Scan(&cols).Error
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cols {
		if c.Nullable != "NO" && !nullableOwnerColumn[c.Table+"."+c.Column] {
			t.Errorf("%s.%s is nullable; owned rows always have an owner", c.Table, c.Column)
		}
		var fks int
		err := db.Raw(`SELECT count(*) FROM pg_constraint k
			JOIN pg_attribute a ON a.attrelid = k.conrelid AND a.attnum = ANY (k.conkey)
			WHERE k.contype = 'f' AND k.conrelid = ('public.' || quote_ident(?))::regclass
			  AND k.confrelid = 'public.users'::regclass AND a.attname = ?`, c.Table, c.Column).Scan(&fks).Error
		if err != nil {
			t.Fatal(err)
		}
		if fks == 0 {
			t.Errorf("%s.%s has no foreign key to users", c.Table, c.Column)
		}
	}
}

// exec runs one statement as u in its own request transaction (a failed statement aborts it).
func exec(h *twousers.Harness, u twousers.User, sql string, args ...any) (int64, error) {
	var n int64
	err := h.Run(u, func(ctx context.Context) error {
		res := database.From(ctx).Exec(sql, args...)
		n = res.RowsAffected
		return res.Error
	})
	return n, err
}

func count(t *testing.T, h *twousers.Harness, u twousers.User, sql string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := h.Run(u, func(ctx context.Context) error {
		return database.From(ctx).Raw(sql, args...).Scan(&n).Error
	}); err != nil {
		t.Fatalf("%s as %s: %v", sql, u.Username, err)
	}
	return n
}

// The patterns in testdata/rls_patterns.sql (copied into DECISIONS.md) do what DECISIONS says.
func TestRLSPatterns(t *testing.T) {
	h := twousers.New(t)
	for _, f := range []string{"testdata/rls_patterns_before.sql", "testdata/rls_patterns.sql"} {
		sql, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if err := h.Owner.Exec(string(sql)).Error; err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
	A, B, admin := h.Alice, h.Bob, h.Admin
	mustOK := func(want int64, u twousers.User, sql string, args ...any) {
		t.Helper()
		n, err := exec(h, u, sql, args...)
		if err != nil || n != want {
			t.Fatalf("%s as %s: %d rows, %v; want %d rows", sql, u.Username, n, err, want)
		}
	}
	mustFail := func(u twousers.User, sql string, args ...any) {
		t.Helper()
		if _, err := exec(h, u, sql, args...); err == nil {
			t.Fatalf("%s as %s succeeded, want an error", sql, u.Username)
		}
	}

	// Backfill gave every existing row to the admin.
	for _, q := range []string{
		"SELECT count(*) FROM example_lists WHERE user_id <> ?",
		"SELECT count(*) FROM example_list_items WHERE user_id <> ?",
		"SELECT count(*) FROM example_foods WHERE created_by <> ?",
	} {
		var n int64
		if err := h.Owner.Raw(q, admin.ID).Scan(&n).Error; err != nil || n != 0 {
			t.Fatalf("%s: %d, %v", q, n, err)
		}
	}

	// Owned parent: per-user uniques, own rows only.
	mustOK(1, A, "INSERT INTO example_lists (name, user_id) VALUES ('groceries', ?)", A.ID)
	mustOK(1, B, "INSERT INTO example_lists (name, user_id) VALUES ('groceries', ?)", B.ID)
	mustFail(A, "INSERT INTO example_lists (name, user_id) VALUES ('groceries', ?)", A.ID)
	mustFail(B, "INSERT INTO example_lists (name, user_id) VALUES ('mine', ?)", A.ID) // WITH CHECK
	if n := count(t, h, B, "SELECT count(*) FROM example_lists"); n != 1 {
		t.Fatalf("Bob sees %d lists, want his 1", n)
	}
	var alices int64
	if err := h.Owner.Raw("SELECT id FROM example_lists WHERE user_id = ?", A.ID).Scan(&alices).Error; err != nil {
		t.Fatal(err)
	}
	mustOK(0, B, "UPDATE example_lists SET name = 'x' WHERE id = ?", alices)
	mustOK(0, B, "DELETE FROM example_lists WHERE id = ?", alices)

	// Owned child: Bob can't hang a row off Alice's parent, as himself (composite FK) or as
	// Alice (WITH CHECK). Alice can.
	mustFail(B, "INSERT INTO example_list_items (list_id, user_id, text) VALUES (?, ?, 'x')", alices, B.ID)
	mustFail(B, "INSERT INTO example_list_items (list_id, user_id, text) VALUES (?, ?, 'x')", alices, A.ID)
	mustOK(1, A, "INSERT INTO example_list_items (list_id, user_id, text) VALUES (?, ?, 'x')", alices, A.ID)
	if n := count(t, h, B, "SELECT count(*) FROM example_list_items"); n != 0 {
		t.Fatalf("Bob sees %d items", n)
	}

	// Catalog: everyone signed in reads; add as yourself; creator or admin edits; admin deletes.
	if n := count(t, h, B, "SELECT count(*) FROM example_foods"); n != 1 {
		t.Fatalf("Bob sees %d foods, want the admin's 1", n)
	}
	if n := count(t, h, twousers.Anonymous, "SELECT count(*) FROM example_foods"); n != 0 {
		t.Fatalf("no identity sees %d foods", n)
	}
	mustOK(1, B, "INSERT INTO example_foods (name, created_by) VALUES ('pear', ?)", B.ID)
	mustFail(B, "INSERT INTO example_foods (name, created_by) VALUES ('plum', ?)", A.ID)
	mustOK(0, A, "UPDATE example_foods SET name = 'x' WHERE name = 'pear'")
	mustOK(1, B, "UPDATE example_foods SET name = 'big pear' WHERE name = 'pear'")
	mustFail(B, "UPDATE example_foods SET created_by = ? WHERE name = 'big pear'", A.ID)
	mustFail(B, "UPDATE example_foods SET deleted_at = now() WHERE name = 'big pear'") // soft delete
	mustOK(0, B, "DELETE FROM example_foods WHERE name = 'big pear'")
	mustOK(1, admin, "UPDATE example_foods SET name = 'pear' WHERE name = 'big pear'")
	mustOK(1, admin, "UPDATE example_foods SET deleted_at = now() WHERE name = 'pear'")
	mustOK(1, admin, "DELETE FROM example_foods WHERE name = 'pear'")
	// The owner role (migrations, seeds) isn't held back by catalog_guard.
	if err := h.Owner.Exec("UPDATE example_foods SET deleted_at = now()").Error; err != nil {
		t.Fatalf("owner soft delete: %v", err)
	}
}

// Advisory locks are per user: Alice and Bob never wait on each other; the same user does.
func TestXactLockPerUser(t *testing.T) {
	h := twousers.New(t)
	ctxA, endA := h.Ctx(h.Alice)
	defer endA(false)
	if err := database.XactLock(ctxA, database.From(ctxA), "workout_logs.date", "2026-01-15"); err != nil {
		t.Fatal(err)
	}
	try := func(u twousers.User) error {
		ctx, end := h.Ctx(u)
		defer end(false)
		db := database.From(ctx)
		if err := db.Exec("SET LOCAL lock_timeout = '200ms'").Error; err != nil {
			return err
		}
		return database.XactLock(ctx, db, "workout_logs.date", "2026-01-15")
	}
	if err := try(h.Bob); err != nil {
		t.Fatalf("Bob waited on Alice's lock: %v", err)
	}
	if err := try(h.Alice); err == nil || !strings.Contains(err.Error(), "lock timeout") {
		t.Fatalf("a second Alice request got the lock while the first held it: %v", err)
	}
}

// A pooled connection that served a request (app.user_id set, then the transaction ended) reads
// the setting as ” afterwards. app_user_id() turns that into NULL: no rows, not an error.
func TestAppUserIDAfterRequest(t *testing.T) {
	h := twousers.New(t)
	sqlDB, err := h.App.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1) // the same session for both
	if err := h.Run(h.Alice, func(ctx context.Context) error {
		return database.From(ctx).Exec("INSERT INTO drink_size_presets (name, amount_oz, user_id, created_at, updated_at) VALUES ('x', 1, ?, now(), now())", h.Alice.ID).Error
	}); err != nil {
		t.Fatal(err)
	}
	var setting string
	var n int64
	if err := h.App.Raw("SELECT current_setting('app.user_id', true)").Scan(&setting).Error; err != nil {
		t.Fatal(err)
	}
	if err := h.App.Raw("SELECT count(*) FROM drink_size_presets").Scan(&n).Error; err != nil || n != 0 {
		t.Fatalf("after the request (setting %q): %d rows, %v; want 0 rows and no error", setting, n, err)
	}
}

// ownerKitVersion is the version of the owner-kit migration (it may be renumbered on rebase).
func ownerKitVersion(t *testing.T) int {
	t.Helper()
	ms, err := migrate.Load(migrations.FS)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range ms {
		if m.Name == "owner_kit" {
			return m.Version
		}
	}
	t.Fatal("no owner_kit migration")
	return 0
}

// migrateTo applies migrations up to target (0 = all) on dsn as the owner.
func migrateTo(t *testing.T, dsn string, target int) error {
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
	_, err = m.Up(context.Background(), target)
	return err
}

// The reference backfill: existing presets (soft-deleted ones too) go to the one admin, and the
// migration stops rather than guess when there isn't exactly one.
func TestOwnerKitMigrationBackfill(t *testing.T) {
	before := ownerKitVersion(t) - 1
	seed := `INSERT INTO users (username, password, email, is_admin, created_at) VALUES
			('owner', 'x', 'o@x', true, now()), ('other', 'x', 'p@x', false, now());
		INSERT INTO drink_size_presets (name, amount_oz, created_at, deleted_at) VALUES
			('Glass', 8, now(), NULL), ('Old', 4, now(), now());`

	dsn := pgtest.NewDatabase(t, "ownerkit_backfill")
	if err := migrateTo(t, dsn, before); err != nil {
		t.Fatal(err)
	}
	db := pgtest.Open(t, dsn)
	if err := db.Exec(seed).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateTo(t, dsn, 0); err != nil {
		t.Fatal(err)
	}
	var wrong int64
	if err := db.Raw(`SELECT count(*) FROM drink_size_presets p
		WHERE p.user_id <> (SELECT id FROM users WHERE username = 'owner')`).Scan(&wrong).Error; err != nil || wrong != 0 {
		t.Fatalf("%d presets not given to the admin (%v)", wrong, err)
	}

	noAdmin := pgtest.NewDatabase(t, "ownerkit_noadmin")
	if err := migrateTo(t, noAdmin, before); err != nil {
		t.Fatal(err)
	}
	if err := pgtest.Open(t, noAdmin).Exec(strings.Replace(seed, "true", "false", 1)).Error; err != nil {
		t.Fatal(err)
	}
	if err := migrateTo(t, noAdmin, 0); err == nil || !strings.Contains(err.Error(), "expected exactly one") {
		t.Fatalf("migration with presets and no admin: %v, want it to stop", err)
	}
}
