package database

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Request code reaches the database only through From(ctx), so every query runs in the request's
// transaction with its identity (app.user_id) set. These patterns are ways around that: a global
// connection, a pool stored in a handler/service, or a context that skips the request transaction.
var requestPathForbidden = []struct {
	re  *regexp.Regexp
	why string
}{
	{regexp.MustCompile(`\bGetDB\(|\bSetDB\(`), "global connection; use database.From(ctx)"},
	{regexp.MustCompile(`(?m)^func conn\(|[^\w.]conn\(\)`), "package-global connection helper; use database.From(ctx)"},
	{regexp.MustCompile(`(?m)^\s+\w+\s+\*gorm\.DB\s*$`), "*gorm.DB stored in a struct; take ctx and use database.From(ctx)"},
	{regexp.MustCompile(`database\.(ConnectApp|ConnectOwner|Open)\(`), "opens its own pool"},
	{regexp.MustCompile(`database\.WithDB\(`), "bypasses the request transaction"},
}

// Code under these paths serves requests.
var requestPaths = []string{"internal/core/", "internal/utils/"}

// Modules still on the package-global conn(); the next PRs of lane B2 convert them and shrink
// this list to nothing.
var notYetConverted = []string{"internal/core/diet/", "internal/core/workout/"}

// Allowed exceptions, each with its reason.
var requestPathAllowed = map[string]string{
	// The session lookup runs on the pool before the request has a user, so the request
	// transaction still begins lazily (DECISIONS.md, B2).
	"internal/core/auth/auth_middleware.go": "database.WithDB",
}

func TestRequestCodeUsesFrom(t *testing.T) {
	root := filepath.Join("..", "..")
	checked := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		inScope := false
		for _, p := range requestPaths {
			inScope = inScope || strings.HasPrefix(rel, p)
		}
		// testutil packages build module test fixtures, not request code.
		if !inScope || strings.Contains(rel, "/testutil/") {
			return nil
		}
		for _, p := range notYetConverted {
			if strings.HasPrefix(rel, p) {
				return nil
			}
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		checked++
		for _, f := range requestPathForbidden {
			for _, loc := range f.re.FindAllIndex(src, -1) {
				match := strings.TrimSpace(string(src[loc[0]:loc[1]]))
				if allowed, ok := requestPathAllowed[rel]; ok && strings.HasPrefix(match, allowed) {
					continue
				}
				line := 1 + strings.Count(string(src[:loc[0]]), "\n")
				t.Errorf("%s:%d: %q: %s", rel, line, match, f.why)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 30 {
		t.Fatalf("only %d files checked; did the layout move?", checked)
	}
}
