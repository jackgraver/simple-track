package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"
)

// lockKey is the pg_advisory_lock key held while migrations run, so two
// concurrent `migrate up` runs (CI + a manual run, two replicas, ...) can't
// both apply the same migration.
const lockKey int64 = 0x73_74_6d_69_67 // "stmig"

const createTableSQL = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version    integer     PRIMARY KEY,
	name       text        NOT NULL,
	checksum   text        NOT NULL,
	applied_at timestamptz NOT NULL DEFAULT now()
)`

// Migrator applies a fixed set of migrations to one database.
type Migrator struct {
	db         *sql.DB
	migrations []Migration
	logf       func(format string, args ...any)
}

// New loads and validates the migrations in fsys. logf receives progress
// messages and warnings; nil discards them.
func New(db *sql.DB, fsys fs.FS, logf func(format string, args ...any)) (*Migrator, error) {
	migrations, err := Load(fsys)
	if err != nil {
		return nil, err
	}
	if logf == nil {
		logf = func(string, ...any) {}
	}
	return &Migrator{db: db, migrations: migrations, logf: logf}, nil
}

// Migrations returns the loaded migrations, ordered by version.
func (m *Migrator) Migrations() []Migration { return m.migrations }

// Latest is the highest version known to this binary (0 if there are none).
func (m *Migrator) Latest() int { return len(m.migrations) }

// Applied is one row of schema_migrations.
type Applied struct {
	Version   int
	Name      string
	Checksum  string
	AppliedAt time.Time
}

// State compares the migration files against the database.
type State struct {
	TableExists bool
	Applied     []Applied   // every row in schema_migrations, by version
	Pending     []Migration // files with no row
	Changed     []Migration // applied, but the file's up section differs from what ran
	Unknown     []Applied   // rows with no file: the DB is ahead of this binary
}

// Current is the highest applied version (0 when nothing is applied).
func (s State) Current() int {
	if len(s.Applied) == 0 {
		return 0
	}
	return s.Applied[len(s.Applied)-1].Version
}

// Status reads schema_migrations without taking the lock or changing anything.
func (m *Migrator) Status(ctx context.Context) (State, error) {
	return m.state(ctx, m.db)
}

// Check is the server's startup check. It returns an error when the database
// is missing migrations this binary needs or an applied file was edited. A
// database that is *ahead* of the binary (e.g. after rolling a deploy back) is
// only logged, so rollbacks aren't blocked.
func (m *Migrator) Check(ctx context.Context) error {
	s, err := m.Status(ctx)
	if err != nil {
		return err
	}
	if !s.TableExists {
		return fmt.Errorf("no schema_migrations table: run `migrate up` (or `migrate baseline` on a database created before migrations)")
	}
	if err := s.changedErr(); err != nil {
		return err
	}
	if len(s.Pending) > 0 {
		return fmt.Errorf("database is at version %04d but this binary needs %04d (%d pending): run `migrate up`",
			s.Current(), m.Latest(), len(s.Pending))
	}
	for _, u := range s.Unknown {
		m.logf("migrate: warning: database has migration %04d_%s which this binary doesn't know (database is ahead)", u.Version, u.Name)
	}
	return nil
}

// Up applies pending migrations in order, up to and including target
// (0 = all). It returns the migrations it applied.
func (m *Migrator) Up(ctx context.Context, target int) ([]Migration, error) {
	if target == 0 {
		target = m.Latest()
	}
	if target < 0 || target > m.Latest() {
		return nil, fmt.Errorf("target version %d out of range (latest is %04d)", target, m.Latest())
	}
	var done []Migration
	err := m.locked(ctx, func(conn *sql.Conn) error {
		s, err := m.state(ctx, conn)
		if err != nil {
			return err
		}
		if len(s.Applied) == 0 {
			if err := m.refuseUntrackedSchema(ctx, conn); err != nil {
				return err
			}
		}
		if err := s.changedErr(); err != nil {
			return err
		}
		for _, p := range s.Pending {
			if p.Version < s.Current() {
				return fmt.Errorf("%s is pending but %04d is already applied; migrations must be applied in order (renumber it after %04d)",
					p.Filename(), s.Current(), s.Current())
			}
		}
		for _, p := range s.Pending {
			if p.Version > target {
				break
			}
			start := time.Now()
			if err := m.run(ctx, conn, p, p.Up, true); err != nil {
				return err
			}
			m.logf("migrate: applied %s (%s)", p.Filename(), time.Since(start).Round(time.Millisecond))
			done = append(done, p)
		}
		return nil
	})
	return done, err
}

// Down rolls back the most recent `steps` applied migrations, newest first.
// Every one of them must exist as a file and be reversible, otherwise nothing
// is rolled back. Checksums aren't verified: rolling back a migration you've
// since edited is exactly the dev "fix it and redo" flow.
func (m *Migrator) Down(ctx context.Context, steps int) ([]Migration, error) {
	if steps < 1 {
		return nil, fmt.Errorf("steps must be at least 1")
	}
	var done []Migration
	err := m.locked(ctx, func(conn *sql.Conn) error {
		s, err := m.state(ctx, conn)
		if err != nil {
			return err
		}
		if steps > len(s.Applied) {
			return fmt.Errorf("asked to roll back %d migration(s) but only %d are applied", steps, len(s.Applied))
		}
		var plan []Migration
		for i := len(s.Applied) - 1; i >= len(s.Applied)-steps; i-- {
			a := s.Applied[i]
			if a.Version > m.Latest() {
				return fmt.Errorf("can't roll back %04d_%s: this binary has no file for it", a.Version, a.Name)
			}
			mig := m.migrations[a.Version-1]
			if mig.Irreversible {
				return fmt.Errorf("can't roll back %s: it is marked irreversible", mig.Filename())
			}
			plan = append(plan, mig)
		}
		for _, mig := range plan {
			start := time.Now()
			if err := m.run(ctx, conn, mig, mig.Down, false); err != nil {
				return err
			}
			m.logf("migrate: rolled back %s (%s)", mig.Filename(), time.Since(start).Round(time.Millisecond))
			done = append(done, mig)
		}
		return nil
	})
	return done, err
}

// Baseline records migrations 1..version as applied without running them.
// It's for adopting a database whose schema was created some other way (the
// old GORM AutoMigrate setup) and only works on an empty history.
func (m *Migrator) Baseline(ctx context.Context, version int) error {
	if version < 1 || version > m.Latest() {
		return fmt.Errorf("baseline version %d out of range (latest is %04d)", version, m.Latest())
	}
	return m.locked(ctx, func(conn *sql.Conn) error {
		s, err := m.state(ctx, conn)
		if err != nil {
			return err
		}
		if len(s.Applied) > 0 {
			return fmt.Errorf("baseline needs an empty migration history; database is already at %04d", s.Current())
		}
		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		for _, mig := range m.migrations[:version] {
			if err := record(ctx, tx, mig, true); err != nil {
				return err
			}
			m.logf("migrate: marked %s as applied (not run)", mig.Filename())
		}
		return tx.Commit()
	})
}

type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (m *Migrator) state(ctx context.Context, q querier) (State, error) {
	var s State
	if err := q.QueryRowContext(ctx, `SELECT to_regclass('schema_migrations') IS NOT NULL`).Scan(&s.TableExists); err != nil {
		return s, fmt.Errorf("check schema_migrations: %w", err)
	}
	applied := map[int]Applied{}
	if s.TableExists {
		rows, err := q.QueryContext(ctx, `SELECT version, name, checksum, applied_at FROM schema_migrations ORDER BY version`)
		if err != nil {
			return s, fmt.Errorf("read schema_migrations: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var a Applied
			if err := rows.Scan(&a.Version, &a.Name, &a.Checksum, &a.AppliedAt); err != nil {
				return s, err
			}
			s.Applied = append(s.Applied, a)
			applied[a.Version] = a
		}
		if err := rows.Err(); err != nil {
			return s, err
		}
	}
	for _, mig := range m.migrations {
		a, ok := applied[mig.Version]
		switch {
		case !ok:
			s.Pending = append(s.Pending, mig)
		case a.Checksum != mig.Checksum:
			s.Changed = append(s.Changed, mig)
		}
	}
	for _, a := range s.Applied {
		if a.Version > m.Latest() {
			s.Unknown = append(s.Unknown, a)
		}
	}
	return s, nil
}

func (s State) changedErr() error {
	if len(s.Changed) == 0 {
		return nil
	}
	names := make([]string, len(s.Changed))
	for i, c := range s.Changed {
		names[i] = c.Filename()
	}
	return fmt.Errorf("applied migration(s) were edited after running: %s; put changes in a new migration (in dev: `migrate redo` if it's the latest)",
		strings.Join(names, ", "))
}

// refuseUntrackedSchema stops `up` from running the baseline over a database
// that already has tables but no history (created by the old AutoMigrate).
func (m *Migrator) refuseUntrackedSchema(ctx context.Context, q querier) error {
	var n int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM pg_tables
		WHERE schemaname = current_schema() AND tablename <> 'schema_migrations'`).Scan(&n)
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("database has %d table(s) but no migration history; if its schema matches %s, run `migrate baseline 1`",
			n, m.migrations[0].Filename())
	}
	return nil
}

// locked runs fn on a dedicated connection holding the advisory lock, after
// making sure schema_migrations exists. Session-level advisory locks belong to
// a connection, hence sql.Conn rather than the pool.
func (m *Migrator) locked(ctx context.Context, fn func(*sql.Conn) error) error {
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var got bool
	if err := conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock($1)`, lockKey).Scan(&got); err != nil {
		return fmt.Errorf("advisory lock: %w", err)
	}
	if !got {
		m.logf("migrate: another migration run holds the lock, waiting...")
		if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, lockKey); err != nil {
			return fmt.Errorf("advisory lock: %w", err)
		}
	}
	defer conn.ExecContext(context.WithoutCancel(ctx), `SELECT pg_advisory_unlock($1)`, lockKey)

	if _, err := conn.ExecContext(ctx, createTableSQL); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}
	return fn(conn)
}

// run executes one direction of a migration and records it. Normally the SQL
// and the schema_migrations change share a transaction, so a migration is
// either fully applied and recorded or not at all.
func (m *Migrator) run(ctx context.Context, conn *sql.Conn, mig Migration, body string, up bool) error {
	dir := "up"
	if !up {
		dir = "down"
	}
	wrap := func(err error) error { return fmt.Errorf("%s (%s): %w", mig.Filename(), dir, err) }

	if mig.NoTransaction {
		// Postgres can't run e.g. CREATE INDEX CONCURRENTLY in a transaction.
		// If the process dies between these two statements the migration ran
		// but isn't recorded, so no-transaction migrations should be written
		// to be re-runnable (IF NOT EXISTS).
		if _, err := conn.ExecContext(ctx, body); err != nil {
			return wrap(err)
		}
		if err := record(ctx, conn, mig, up); err != nil {
			return wrap(err)
		}
		return nil
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return wrap(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, body); err != nil {
		return wrap(err)
	}
	if err := record(ctx, tx, mig, up); err != nil {
		return wrap(err)
	}
	if err := tx.Commit(); err != nil {
		return wrap(err)
	}
	return nil
}

func record(ctx context.Context, e execer, mig Migration, up bool) error {
	var err error
	if up {
		_, err = e.ExecContext(ctx, `INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)`,
			mig.Version, mig.Name, mig.Checksum)
	} else {
		var res sql.Result
		res, err = e.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = $1`, mig.Version)
		if err == nil {
			if n, _ := res.RowsAffected(); n != 1 {
				err = errors.New("schema_migrations row was not found")
			}
		}
	}
	if err != nil {
		return fmt.Errorf("record in schema_migrations: %w", err)
	}
	return nil
}
