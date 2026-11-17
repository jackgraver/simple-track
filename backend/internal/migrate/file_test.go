package migrate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"be-simpletracker/migrations"
)

func TestParse(t *testing.T) {
	m, err := Parse("0003_add_email.sql", "-- why: login by email\r\n-- migrate:up\r\nALTER TABLE users ADD email text;\r\n\r\n-- migrate:down\r\nALTER TABLE users DROP email;\r\n")
	if err != nil {
		t.Fatal(err)
	}
	if m.Version != 3 || m.Name != "add_email" || m.Filename() != "0003_add_email.sql" {
		t.Errorf("got version %d name %q", m.Version, m.Name)
	}
	if m.Up != "ALTER TABLE users ADD email text;" || m.Down != "ALTER TABLE users DROP email;" {
		t.Errorf("sections: up %q down %q", m.Up, m.Down)
	}
	if m.Irreversible || m.NoTransaction {
		t.Error("unexpected flags")
	}

	// CRLF and LF versions must have the same checksum.
	lf, _ := Parse("0003_add_email.sql", "-- migrate:up\nALTER TABLE users ADD email text;\n-- migrate:down\nALTER TABLE users DROP email;\n")
	if lf.Checksum != m.Checksum {
		t.Error("checksum depends on line endings")
	}
	// Editing only the down section keeps the checksum.
	edited, _ := Parse("0003_add_email.sql", "-- migrate:up\nALTER TABLE users ADD email text;\n-- migrate:down\nALTER TABLE users DROP COLUMN email;\n")
	if edited.Checksum != m.Checksum {
		t.Error("checksum covers the down section")
	}
}

func TestParseFlags(t *testing.T) {
	m, err := Parse("0001_idx.sql", "-- migrate:no-transaction\n-- migrate:up\nCREATE INDEX CONCURRENTLY i ON t (c);\n-- migrate:down irreversible\n-- can't, sorry\n")
	if err != nil {
		t.Fatal(err)
	}
	if !m.NoTransaction || !m.Irreversible {
		t.Errorf("flags: no-tx %v irreversible %v", m.NoTransaction, m.Irreversible)
	}
}

func TestParseErrors(t *testing.T) {
	cases := map[string]struct{ file, content, want string }{
		"bad name":          {"1_x.sql", "", "file name"},
		"version zero":      {"0000_x.sql", "", "start at 0001"},
		"uppercase name":    {"0001_Add.sql", "", "file name"},
		"no up":             {"0001_x.sql", "-- hello\n", "missing '-- migrate:up'"},
		"sql before up":     {"0001_x.sql", "SELECT 1;\n-- migrate:up\nSELECT 1;\n-- migrate:down\nSELECT 1;", "SQL before"},
		"no down":           {"0001_x.sql", "-- migrate:up\nSELECT 1;\n", "missing '-- migrate:down'"},
		"empty up":          {"0001_x.sql", "-- migrate:up\n-- todo\n-- migrate:down\nSELECT 1;", "up section is empty"},
		"empty down":        {"0001_x.sql", "-- migrate:up\nSELECT 1;\n-- migrate:down\n", "down section is empty"},
		"irreversible+sql":  {"0001_x.sql", "-- migrate:up\nSELECT 1;\n-- migrate:down irreversible\nSELECT 1;", "irreversible but contains SQL"},
		"down before up":    {"0001_x.sql", "-- migrate:down\nSELECT 1;\n-- migrate:up\nSELECT 1;", "after '-- migrate:up'"},
		"duplicate up":      {"0001_x.sql", "-- migrate:up\nSELECT 1;\n-- migrate:up\n", "duplicate"},
		"late no-tx":        {"0001_x.sql", "-- migrate:up\n-- migrate:no-transaction\nSELECT 1;\n-- migrate:down\nSELECT 1;", "must come before"},
		"typo in directive": {"0001_x.sql", "-- migrate:upp\nSELECT 1;", "unknown directive"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(c.file, c.content)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want error containing %q", err, c.want)
			}
		})
	}
}

const ok = "-- migrate:up\nSELECT 1;\n-- migrate:down\nSELECT 1;\n"

func TestLoadOrdersAndValidates(t *testing.T) {
	ms, err := Load(fstest.MapFS{
		"0002_b.sql": {Data: []byte(ok)},
		"0001_a.sql": {Data: []byte(ok)},
		"README.md":  {Data: []byte("ignored")},
		"migrate.go": {Data: []byte("ignored")},
		"0003_c.sql": {Data: []byte(ok)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 3 || ms[0].Name != "a" || ms[2].Name != "c" {
		t.Fatalf("got %+v", ms)
	}

	if _, err := Load(fstest.MapFS{"0001_a.sql": {Data: []byte(ok)}, "0003_c.sql": {Data: []byte(ok)}}); err == nil || !strings.Contains(err.Error(), "gaps") {
		t.Errorf("gap: got %v", err)
	}
	if _, err := Load(fstest.MapFS{"0001_a.sql": {Data: []byte(ok)}, "0001_b.sql": {Data: []byte(ok)}}); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("duplicate: got %v", err)
	}
	if _, err := Load(fstest.MapFS{"001_a.sql": {Data: []byte(ok)}}); err == nil {
		t.Error("badly named .sql file was accepted")
	}
}

// The real migrations must always parse; this is what the server embeds.
func TestEmbeddedMigrationsParse(t *testing.T) {
	ms, err := Load(migrations.FS)
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) == 0 {
		t.Fatal("no migrations embedded")
	}
}

func TestNewFile(t *testing.T) {
	dir := t.TempDir()
	// An unfinished migration still reserves its number.
	os.WriteFile(filepath.Join(dir, "0001_first.sql"), []byte(ok), 0o644)
	os.WriteFile(filepath.Join(dir, "0002_wip.sql"), []byte("-- migrate:up\n"), 0o644)

	path, err := NewFile(dir, "Add user Email!")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "0003_add_user_email.sql" {
		t.Errorf("got %s", path)
	}
	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "-- migrate:up") || !strings.Contains(string(content), "-- migrate:down") {
		t.Errorf("template: %q", content)
	}
	if _, err := NewFile(dir, "!!!"); err == nil {
		t.Error("empty slug accepted")
	}
}
