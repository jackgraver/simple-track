package grocery

import (
	"reflect"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestParseList(t *testing.T) {
	text := "Hey can you grab:\n- milk\n* 12 eggs, bread; Bread\n1. bananas\n2) [ ] Peanut butter.\n\n•  avocados  \n"
	got := ParseList(text)
	want := []string{"Hey can you grab", "milk", "12 eggs", "bread", "bananas", "Peanut butter", "avocados"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseList:\n got %q\nwant %q", got, want)
	}
}

func TestMatchKey(t *testing.T) {
	cases := map[string]string{
		"eggs":               "egg",
		"Eggs":               "egg",
		"5 apples":           "apple",
		"12 Eggs":            "egg",
		"ground beef":        "ground beef",
		"potatoes":           "potato",
		"Strawberries":       "strawberry",
		"hummus":             "hummus",
		"2% milk":            "2% milk",
		"7up":                "7up",
		"Peanut Butter (PB)": "peanut butter pb",
	}
	for in, want := range cases {
		if got := matchKey(in); got != want {
			t.Errorf("matchKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMatchLines(t *testing.T) {
	known := []KnownItem{
		{Name: "Eggs", Count: 10},
		{Name: "Banana", Count: 6},
		{Name: "Ground Beef", Count: 3},
		{Name: "Almond Milk", Count: 2},
		{Name: "Milk", Count: 8},
	}
	lines := []string{"12 eggs", "bannanas", "ground beef", "milk", "dragonfruit", "egg"}
	got := MatchLines(lines, known, []string{"Milk"})
	want := []ParsedLine{
		{Input: "12 eggs", Name: "Eggs", Matched: true},
		{Input: "bannanas", Name: "Banana", Matched: true},
		{Input: "ground beef", Name: "Ground Beef", Matched: true},
		{Input: "milk", Name: "Milk", Matched: true, OnList: true},
		{Input: "dragonfruit", Name: "dragonfruit"},
		{Input: "egg", Name: "Eggs", Matched: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MatchLines:\n got %+v\nwant %+v", got, want)
	}
}

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&GroceryItem{}); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestParseTextListAndCreateItems(t *testing.T) {
	db := testDB(t)
	now := time.Now()
	for _, n := range []string{"Eggs", "Eggs", "Bread"} {
		db.Create(&GroceryItem{Name: n, CompletedAt: &now})
	}
	db.Create(&GroceryItem{Name: "Bread"}) // active

	lines, err := ParseTextList(db, "eggs\nbread\ncheese")
	if err != nil {
		t.Fatal(err)
	}
	want := []ParsedLine{
		{Input: "eggs", Name: "Eggs", Matched: true},
		{Input: "bread", Name: "Bread", Matched: true, OnList: true},
		{Input: "cheese", Name: "cheese"},
	}
	if !reflect.DeepEqual(lines, want) {
		t.Fatalf("ParseTextList:\n got %+v\nwant %+v", lines, want)
	}

	created, skipped, err := CreateItems(db, []string{"Eggs", "bread", " cheese ", "Cheese", ""})
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 2 || created[0].Name != "Eggs" || created[1].Name != "cheese" {
		t.Fatalf("created = %+v", created)
	}
	if !reflect.DeepEqual(skipped, []string{"bread", "Cheese"}) {
		t.Fatalf("skipped = %q", skipped)
	}
	active, _ := ListActiveItems(db)
	if len(active) != 3 {
		t.Fatalf("active = %d, want 3", len(active))
	}
}
