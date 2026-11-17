// Package migrate applies versioned SQL migrations to Postgres.
//
// Migrations are files named NNNN_snake_name.sql (versions 1..N, no gaps):
//
//	-- migrate:no-transaction        (optional, must come before up)
//	-- migrate:up
//	CREATE TABLE ...;
//
//	-- migrate:down                  (or: -- migrate:down irreversible)
//	DROP TABLE ...;
//
// Applied versions are recorded in the schema_migrations table together with a
// checksum of the up section, so an applied migration that is later edited is
// detected instead of silently diverging from the database.
package migrate

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Migration is one parsed migration file.
type Migration struct {
	Version       int
	Name          string // snake_case part of the filename
	Up            string
	Down          string
	Irreversible  bool // down section is marked irreversible; rollback is refused
	NoTransaction bool // run outside a transaction (e.g. CREATE INDEX CONCURRENTLY)
	Checksum      string
}

// Filename returns the file name the migration was loaded from.
func (m Migration) Filename() string { return fmt.Sprintf("%04d_%s.sql", m.Version, m.Name) }

var (
	fileRe      = regexp.MustCompile(`^(\d{4})_([a-z0-9]+(?:_[a-z0-9]+)*)\.sql$`)
	directiveRe = regexp.MustCompile(`^--\s*migrate:(\S+)\s*(.*?)\s*$`)
)

// Load reads and validates every .sql file at the root of fsys. Versions must
// run 1..N without gaps or duplicates.
func Load(fsys fs.FS) ([]Migration, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	var migrations []Migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		content, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		m, err := Parse(e.Name(), string(content))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].Version < migrations[j].Version })
	for i, m := range migrations {
		if want := i + 1; m.Version != want {
			if i > 0 && m.Version == migrations[i-1].Version {
				return nil, fmt.Errorf("duplicate migration version %04d: %s and %s", m.Version, migrations[i-1].Filename(), m.Filename())
			}
			return nil, fmt.Errorf("migration versions must be 1..N without gaps: expected %04d, found %s", want, m.Filename())
		}
	}
	return migrations, nil
}

// Parse parses a single migration file. filename must match NNNN_snake_name.sql.
func Parse(filename, content string) (Migration, error) {
	fail := func(format string, args ...any) (Migration, error) {
		return Migration{}, fmt.Errorf("%s: %s", filename, fmt.Sprintf(format, args...))
	}
	match := fileRe.FindStringSubmatch(filename)
	if match == nil {
		return fail("file name must look like 0001_snake_case_name.sql")
	}
	version, _ := strconv.Atoi(match[1])
	if version == 0 {
		return fail("versions start at 0001")
	}
	m := Migration{Version: version, Name: match[2]}

	// Normalise line endings so a checkout with CRLF doesn't change checksums.
	content = strings.ReplaceAll(content, "\r\n", "\n")

	const (
		header = iota
		up
		down
	)
	section := header
	var upLines, downLines []string
	for i, line := range strings.Split(content, "\n") {
		d := directiveRe.FindStringSubmatch(strings.TrimSpace(line))
		if d == nil {
			switch section {
			case header:
				if isSQL(line) {
					return fail("line %d: SQL before '-- migrate:up'", i+1)
				}
			case up:
				upLines = append(upLines, line)
			case down:
				downLines = append(downLines, line)
			}
			continue
		}
		name, arg := d[1], d[2]
		switch {
		case name == "no-transaction" && arg == "":
			if section != header {
				return fail("line %d: '-- migrate:no-transaction' must come before '-- migrate:up'", i+1)
			}
			m.NoTransaction = true
		case name == "up" && arg == "":
			if section != header {
				return fail("line %d: duplicate or misplaced '-- migrate:up'", i+1)
			}
			section = up
		case name == "down" && (arg == "" || arg == "irreversible"):
			if section != up {
				return fail("line %d: '-- migrate:down' must come once, after '-- migrate:up'", i+1)
			}
			section = down
			m.Irreversible = arg == "irreversible"
		default:
			return fail("line %d: unknown directive %q (expected up, down, down irreversible or no-transaction)", i+1, strings.TrimSpace(line))
		}
	}

	if section == header {
		return fail("missing '-- migrate:up'")
	}
	if section == up {
		return fail("missing '-- migrate:down' (use '-- migrate:down irreversible' if it can't be undone)")
	}
	m.Up = strings.TrimSpace(strings.Join(upLines, "\n"))
	m.Down = strings.TrimSpace(strings.Join(downLines, "\n"))
	if !isSQL(m.Up) {
		return fail("up section is empty")
	}
	if m.Irreversible && isSQL(m.Down) {
		return fail("down section is marked irreversible but contains SQL")
	}
	if !m.Irreversible && !isSQL(m.Down) {
		return fail("down section is empty; write the rollback or mark it '-- migrate:down irreversible'")
	}
	sum := sha256.Sum256([]byte(m.Up))
	m.Checksum = hex.EncodeToString(sum[:])
	return m, nil
}

// isSQL reports whether s contains anything besides whitespace and -- comments.
func isSQL(s string) bool {
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			return true
		}
	}
	return false
}

const newFileTemplate = `-- migrate:up


-- migrate:down

`

// NewFile creates the next numbered migration in dir and returns its path.
// name is converted to snake_case ("Add user email" -> add_user_email).
func NewFile(dir, name string) (string, error) {
	slug := strings.Trim(regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(strings.ToLower(name), "_"), "_")
	if slug == "" {
		return "", fmt.Errorf("migration name %q has no usable characters", name)
	}
	// Count existing files by name only: a half-written migration (which
	// wouldn't parse yet) must still reserve its version number.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read migrations dir: %w", err)
	}
	next := 1
	for _, e := range entries {
		if match := fileRe.FindStringSubmatch(e.Name()); match != nil {
			if v, _ := strconv.Atoi(match[1]); v >= next {
				next = v + 1
			}
		}
	}
	path := filepath.Join(dir, fmt.Sprintf("%04d_%s.sql", next, slug))
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.WriteString(newFileTemplate); err != nil {
		return "", err
	}
	return path, nil
}
