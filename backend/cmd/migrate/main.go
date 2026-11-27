// Command migrate applies the SQL migrations in backend/migrations.
//
//	migrate up [VERSION]     apply pending migrations (up to VERSION)
//	migrate down [STEPS]     roll back the latest STEPS migrations (default 1)
//	migrate redo             roll back and re-apply the latest migration (dev)
//	migrate status           list applied and pending migrations
//	migrate check            exit 1 unless the database is fully migrated
//	migrate baseline VERSION mark 1..VERSION applied without running them
//	migrate new NAME         create the next migration file in -dir
//
// The migrations are embedded at build time, so a deployed binary always
// applies exactly the set it shipped with. The database is chosen the same way
// as the server's (DATABASE_URL, or the APP_ENV-specific variable).
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"be-simpletracker/internal/database"
	"be-simpletracker/internal/migrate"
	"be-simpletracker/migrations"
)

const usage = `usage: migrate [flags] <command> [arg]

commands:
  up [VERSION]       apply pending migrations (all, or up to VERSION)
  down [STEPS]       roll back the latest STEPS migrations (default 1)
  redo               roll back and re-apply the latest migration
  status             list applied and pending migrations
  check              exit 1 unless the database is fully migrated
  baseline VERSION   mark 1..VERSION as applied without running them
  new NAME           create the next migration file in -dir

flags:
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	dir := flag.String("dir", "migrations", "migrations directory (only used by `new`)")
	timeout := flag.Duration("timeout", 10*time.Minute, "give up after this long")
	flag.Usage = func() {
		fmt.Fprint(flag.CommandLine.Output(), usage)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() < 1 || flag.NArg() > 2 {
		flag.Usage()
		os.Exit(2)
	}
	cmd, arg := flag.Arg(0), flag.Arg(1)

	if cmd == "new" {
		if arg == "" {
			return errors.New("usage: migrate new NAME")
		}
		path, err := migrate.NewFile(*dir, arg)
		if err != nil {
			return err
		}
		fmt.Println("created", path)
		return nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx, cancelTimeout := context.WithTimeout(ctx, *timeout)
	defer cancelTimeout()

	gdb, err := database.ConnectOwner()
	if err != nil {
		return err
	}
	db, err := gdb.DB()
	if err != nil {
		return err
	}
	defer db.Close()

	m, err := migrate.New(db, migrations.FS, func(format string, args ...any) {
		fmt.Printf(format+"\n", args...)
	})
	if err != nil {
		return err
	}

	switch cmd {
	case "up":
		target, err := optionalInt(arg, 0)
		if err != nil {
			return err
		}
		done, err := m.Up(ctx, target)
		if err != nil {
			return err
		}
		if len(done) == 0 {
			fmt.Println("migrate: already up to date")
		}
		if target != 0 && target < m.Latest() {
			return nil // the app role may not exist yet at an older version
		}
		return ensureAppLogin(ctx, db)
	case "down":
		steps, err := optionalInt(arg, 1)
		if err != nil {
			return err
		}
		_, err = m.Down(ctx, steps)
		return err
	case "redo":
		done, err := m.Down(ctx, 1)
		if err != nil {
			return err
		}
		_, err = m.Up(ctx, done[0].Version)
		return err
	case "status":
		return printStatus(ctx, m)
	case "check":
		return m.Check(ctx)
	case "baseline":
		version, err := optionalInt(arg, 0)
		if err != nil || version == 0 {
			return errors.New("usage: migrate baseline VERSION")
		}
		return m.Baseline(ctx, version)
	default:
		flag.Usage()
		os.Exit(2)
	}
	return nil
}

// ensureAppLogin lets the server's role (database.AppRole) log in with the password in the app
// DSN (DATABASE_URL / DATABASE_URL_PRODUCTION). The migrations create the role without a
// password, so this is where the secret is applied, from the environment only. A no-op when the
// app DSN already connects or names another role.
func ensureAppLogin(ctx context.Context, owner *sql.DB) error {
	appDSN, err := database.AppDSN()
	if err != nil {
		return err
	}
	changed, err := database.EnsureAppLogin(ctx, owner, appDSN)
	if err != nil {
		return fmt.Errorf("app role: %w", err)
	}
	if changed {
		fmt.Printf("migrate: set login password for %s from the app DSN\n", database.AppRole)
	}
	return nil
}

func optionalInt(s string, def int) (int, error) {
	if s == "" {
		return def, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("expected a non-negative number, got %q", s)
	}
	return n, nil
}

func printStatus(ctx context.Context, m *migrate.Migrator) error {
	s, err := m.Status(ctx)
	if err != nil {
		return err
	}
	applied := map[int]migrate.Applied{}
	for _, a := range s.Applied {
		applied[a.Version] = a
	}
	changed := map[int]bool{}
	for _, c := range s.Changed {
		changed[c.Version] = true
	}
	for _, mig := range m.Migrations() {
		a, ok := applied[mig.Version]
		switch {
		case !ok:
			fmt.Printf("  pending   %s\n", mig.Filename())
		case changed[mig.Version]:
			fmt.Printf("  CHANGED   %s  (applied %s, file edited since)\n", mig.Filename(), a.AppliedAt.Local().Format(time.DateTime))
		default:
			fmt.Printf("  applied   %s  %s\n", mig.Filename(), a.AppliedAt.Local().Format(time.DateTime))
		}
	}
	for _, u := range s.Unknown {
		fmt.Printf("  UNKNOWN   %04d_%s.sql  (in database, no file in this binary)\n", u.Version, u.Name)
	}
	fmt.Printf("\ndatabase at %04d, latest %04d, %d pending\n", s.Current(), m.Latest(), len(s.Pending))
	return nil
}
