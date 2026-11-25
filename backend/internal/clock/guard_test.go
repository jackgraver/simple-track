package clock

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Domain code must read the time through clock.Now so the e2e suite can move it. These paths
// use time.Now on purpose: durations, benchmarks, uniqueness nonces, and offline tools.
var timeNowAllowed = []string{
	"internal/clock/",
	"internal/migrate/",
	"internal/utils/benchmarker.go",
	"internal/core/diet/controller/meal_route.go", // nanosecond suffix that keeps quick-log food names unique
	"cmd/",
}

func TestNoDirectTimeNow(t *testing.T) {
	root := filepath.Join("..", "..")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		for _, a := range timeNowAllowed {
			if strings.HasPrefix(rel, a) {
				return nil
			}
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(src), "time.Now()") {
			t.Errorf("%s calls time.Now(); use clock.Now() (or add it to timeNowAllowed with a reason)", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
