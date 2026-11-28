package apierr

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// BACKLOG #6: a 500 never carries the raw error (SQL, constraint and policy names). Use
// apierr.Internal, which logs it and answers a fixed message.
var rawInternalError = regexp.MustCompile(`(StatusInternalServerError|[^0-9]500)[^\n]*\.Error\(\)`)

// Sites that predate the rule, as a count per file. Lanes move their handlers to apierr and lower
// or delete their entries; a file may go below its count, never above.
var rawInternalErrorAllowed = map[string]int{
	"internal/core/diet/controller/diet_log_route.go":       6,  // C2
	"internal/core/diet/controller/diet_plan_route.go":      2,  // C2
	"internal/core/media/controller/media_route.go":         1,  // C5
	"internal/core/money/controller/routes.go":              2,  // C4
	"internal/core/tracking/grocery/routes.go":              4,  // C1
	"internal/core/tracking/missed/routes.go":               1,  // C1
	"internal/core/tracking/profile/routes.go":              1,  // C1
	"internal/core/tracking/steps/routes.go":                1,  // C1
	"internal/core/tracking/water/routes.go":                1,  // C1 (water logs; presets done in B3)
	"internal/core/tracking/weight/routes.go":               2,  // C1
	"internal/core/workout/controller/exercises_route.go":   10, // C3
	"internal/core/workout/controller/workoutlog_route.go":  4,  // C3
	"internal/core/workout/controller/workoutplan_route.go": 11, // C3
}

func TestNoRawErrorsIn500s(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	sites := map[string][]string{}
	checked := 0
	err := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if !strings.HasPrefix(rel, "internal/core/") && !strings.HasPrefix(rel, "internal/utils/") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		checked++
		for _, loc := range rawInternalError.FindAllIndex(src, -1) {
			line := 1 + strings.Count(string(src[:loc[0]]), "\n")
			sites[rel] = append(sites[rel], rel+":"+strconv.Itoa(line))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if checked < 50 {
		t.Fatalf("only %d files checked; did the layout move?", checked)
	}
	files := make([]string, 0, len(sites))
	for f := range sites {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		if got, max := len(sites[f]), rawInternalErrorAllowed[f]; got > max {
			t.Errorf("%d new 500 response(s) with a raw err.Error() (allowlist has %d for %s); use apierr.Internal (BACKLOG #6):\n\t%s",
				got-max, max, f, strings.Join(sites[f], "\n\t"))
		}
	}
}

func TestRawInternalErrorPattern(t *testing.T) {
	for src, want := range map[string]bool{
		`c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})`:      true,
		`c.JSON(500, gin.H{"error": "x: " + err.Error()})`:                         true,
		`c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})`:               false,
		`apierr.Internal(c, err)`:                                                  false,
		`c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})`: false,
	} {
		if got := rawInternalError.MatchString(src); got != want {
			t.Errorf("%s: match %v, want %v", src, got, want)
		}
	}
}
