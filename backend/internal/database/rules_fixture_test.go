package database

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// The rule checker on small fixtures, so TestOwnerKitRules passing means something.

func fixture(t *testing.T, srcs map[string]string) ruleResult {
	t.Helper()
	fset := token.NewFileSet()
	var files []goFile
	for name, src := range srcs {
		f, err := parser.ParseFile(fset, name, src, parser.ParseComments)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, goFile{rel: name, src: []byte(src), file: f, fset: fset})
	}
	return checkOwnerKitRules(files)
}

const fixtureModels = `package models
import "be-simpletracker/internal/database"
type Item struct { database.Owner; Name string }
type Food struct { database.Creator; Name string }
type Wrapper struct { Items []*Item }
type Plain struct { Name string }
`

func TestRulesFixture_binds(t *testing.T) {
	r := fixture(t, map[string]string{
		"models.go": fixtureModels,
		"routes.go": `package routes
import ("os"; "models"; "gorm.io/gorm/clause")
type itemsBody struct { Items []models.Item }
type okBody struct { Name string }
func a(c Ctx) { var it models.Item; c.ShouldBindJSON(&it) }
func b(c Ctx) { body := itemsBody{}; c.ShouldBindJSON(&body) }
func d(c Ctx) { w := &models.Wrapper{}; c.BindJSON(w) }
func e(c Ctx) { c.ShouldBindJSON(target()) }
func f(c Ctx) { var body okBody; c.ShouldBindJSON(&body); var p models.Plain; c.ShouldBind(&p) }
func g(c Ctx) { var body struct{ CreatedBy uint }; c.ShouldBindJSON(&body) }
func h(c Ctx, f *models.Food) { c.ShouldBindJSON(f) }
func w(db DB) {
	db.Create(&x)
	db.Omit(clause.Associations).Create(&x)
	db.WithContext(ctx).Omit(clause.Associations).Where("a").Save(&x)
	os.Create("f")
	row.UserID = 3
	db.Updates(map[string]any{"created_by": 1})
	db.Exec("SELECT pg_advisory_xact_lock(1)")
}
func (i *Item) BeforeCreate(tx *DB) error { return nil }
`,
	})
	want := []string{"routes.go:5:", "routes.go:6:", "routes.go:7:", "routes.go:8:", "routes.go:10:", "routes.go:11:", "routes.go:21:"}
	if len(r.errors) != len(want) {
		t.Fatalf("errors:\n%s", strings.Join(r.errors, "\n"))
	}
	for i, w := range want {
		if !strings.HasPrefix(r.errors[i], w) {
			t.Errorf("error %d = %q, want %s...", i, r.errors[i], w)
		}
	}
	if n := len(r.createWithAssoc["routes.go"]); n != 1 {
		t.Errorf("creates without Omit: %v", r.createWithAssoc)
	}
	if n := len(r.ownerWrites["routes.go"]); n != 2 {
		t.Errorf("owner writes: %v", r.ownerWrites)
	}
	if n := len(r.rawLocks["routes.go"]); n != 1 {
		t.Errorf("raw locks: %v", r.rawLocks)
	}
}

func TestRulesFixture_ownerAndCreator(t *testing.T) {
	r := fixture(t, map[string]string{
		"models.go": "package models\nimport \"be-simpletracker/internal/database\"\ntype Both struct { database.Owner; database.Creator }\n",
	})
	if len(r.errors) != 1 || !strings.Contains(r.errors[0], "embeds both") {
		t.Fatalf("errors: %v", r.errors)
	}
}

func TestRulesFixture_beforeCreateThatCallsOwner(t *testing.T) {
	r := fixture(t, map[string]string{
		"models.go": fixtureModels + `func (i *Item) BeforeCreate(tx *DB) error { return i.Owner.BeforeCreate(tx) }`,
	})
	if len(r.errors) != 0 {
		t.Fatalf("errors: %v", r.errors)
	}
}
