package service

import (
	"be-simpletracker/internal/database"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"be-simpletracker/internal/core/media/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.MediaItem{}, &models.UserMedia{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// ctxFor is a context acting as user id with no request transaction (module tests on SQLite).
func ctxFor(db *gorm.DB, id uint) context.Context {
	return database.WithIdentity(database.WithDB(context.Background(), db), database.Identity{UserID: id})
}

func decode(t *testing.T, raw string) MediaInput {
	t.Helper()
	var in MediaInput
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatal(err)
	}
	return in
}

func decodePersonal(t *testing.T, raw string) PersonalInput {
	t.Helper()
	var in PersonalInput
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatal(err)
	}
	return in
}

func decodeCatalog(t *testing.T, raw string) CatalogInput {
	t.Helper()
	var in CatalogInput
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatal(err)
	}
	return in
}

// create is CreateMedia without the created flag, for tests that don't care.
func create(ctx context.Context, in MediaInput) (*Entry, error) {
	e, _, err := CreateMedia(ctx, in, now)
	return e, err
}

var now = time.Date(2026, time.November, 15, 14, 30, 0, 0, time.Local)

func TestCreateDefaultsToPlannedAndRequiresTypeAndTitle(t *testing.T) {
	db := setupTestDB(t)
	item, err := create(ctxFor(db, 1), decode(t, `{"type":"movie","title":"  Heat  ","release_year":1995}`))
	if err != nil {
		t.Fatal(err)
	}
	if item.Status != models.MediaStatusPlanned || item.Title != "Heat" || *item.ReleaseYear != 1995 {
		t.Fatalf("unexpected item: %+v", item)
	}
	for _, raw := range []string{`{"title":"x"}`, `{"type":"movie"}`, `{"type":"vhs","title":"x"}`, `{"type":"movie","title":"   "}`} {
		if _, err := create(ctxFor(db, 1), decode(t, raw)); !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: expected ErrInvalid, got %v", raw, err)
		}
	}
}

func TestRatingValidation(t *testing.T) {
	db := setupTestDB(t)
	for raw, ok := range map[string]bool{
		`0.5`: true, `3`: true, `4.5`: true, `5`: true,
		`0`: false, `5.5`: false, `3.3`: false, `-1`: false,
	} {
		_, err := create(ctxFor(db, 1), decode(t, `{"type":"book","title":"x","rating":`+raw+`}`))
		if ok && err != nil {
			t.Errorf("rating %s: unexpected error %v", raw, err)
		}
		if !ok && !errors.Is(err, ErrInvalid) {
			t.Errorf("rating %s: expected ErrInvalid, got %v", raw, err)
		}
	}
}

func TestPatchOmittedVsNull(t *testing.T) {
	db := setupTestDB(t)
	item, err := create(ctxFor(db, 1), decode(t, `{"type":"movie","title":"Heat","rating":4.5,"cover_url":"http://x/y.jpg"}`))
	if err != nil {
		t.Fatal(err)
	}
	item, err = UpdateMedia(ctxFor(db, 1), item.ID, decodePersonal(t, `{"review":"great"}`), now)
	if err != nil {
		t.Fatal(err)
	}
	if item.Rating == nil || *item.Rating != 4.5 || item.CoverURL == "" || item.Review != "great" {
		t.Fatalf("omitted fields should be untouched: %+v", item)
	}
	item, err = UpdateMedia(ctxFor(db, 1), item.ID, decodePersonal(t, `{"rating":null}`), now)
	if err != nil {
		t.Fatal(err)
	}
	cat, err := UpdateCatalog(ctxFor(db, 1), item.MediaItemID, decodeCatalog(t, `{"cover_url":null}`))
	if err != nil {
		t.Fatal(err)
	}
	if item.Rating != nil || cat.CoverURL != "" {
		t.Fatalf("null should clear fields: %+v", item)
	}
	if _, err := UpdateCatalog(ctxFor(db, 1), item.MediaItemID, decodeCatalog(t, `{"title":null}`)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid for null title, got %v", err)
	}
}

func TestStatusTransitionsFillDates(t *testing.T) {
	db := setupTestDB(t)
	today := time.Date(2026, time.November, 15, 0, 0, 0, 0, time.Local)
	item, err := create(ctxFor(db, 1), decode(t, `{"type":"book","title":"Dune","length":600}`))
	if err != nil {
		t.Fatal(err)
	}
	item, err = UpdateMedia(ctxFor(db, 1), item.ID, decodePersonal(t, `{"status":"in_progress"}`), now)
	if err != nil {
		t.Fatal(err)
	}
	if item.StartedAt == nil || !item.StartedAt.Equal(today) {
		t.Fatalf("expected started_at today, got %v", item.StartedAt)
	}
	item, err = UpdateMedia(ctxFor(db, 1), item.ID, decodePersonal(t, `{"status":"completed"}`), now)
	if err != nil {
		t.Fatal(err)
	}
	if item.FinishedAt == nil || !item.FinishedAt.Equal(today) || item.TimesConsumed != 1 || item.Progress == nil || *item.Progress != 600 {
		t.Fatalf("completion defaults not applied: %+v", item)
	}

	// Explicit dates win over defaults.
	other, err := create(ctxFor(db, 1), decode(t, `{"type":"movie","title":"Alien","status":"completed","finished_at":"2020-05-01"}`))
	if err != nil {
		t.Fatal(err)
	}
	if other.FinishedAt.Year() != 2020 {
		t.Fatalf("explicit finished_at overwritten: %v", other.FinishedAt)
	}

	if _, err := UpdateMedia(ctxFor(db, 1), other.ID, decodePersonal(t, `{"started_at":"2021-01-01"}`), now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected finished-before-started error, got %v", err)
	}
	if _, err := UpdateMedia(ctxFor(db, 1), other.ID, decodePersonal(t, `{"started_at":"01/02/2020"}`), now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected bad date error, got %v", err)
	}
}

func TestListFiltersAndSorts(t *testing.T) {
	db := setupTestDB(t)
	for _, raw := range []string{
		`{"type":"movie","title":"Heat","creator":"Michael Mann","status":"completed","rating":4.5,"favorite":true}`,
		`{"type":"movie","title":"Alien","creator":"Ridley Scott","status":"completed","rating":5}`,
		`{"type":"book","title":"Dune","creator":"Frank Herbert"}`,
		`{"type":"tv","title":"The Wire","status":"in_progress"}`,
	} {
		if _, err := create(ctxFor(db, 1), decode(t, raw)); err != nil {
			t.Fatal(err)
		}
	}
	titles := func(items []Entry) []string {
		out := []string{}
		for _, i := range items {
			out = append(out, i.Title)
		}
		return out
	}
	cases := []struct {
		name   string
		filter ListFilter
		want   []string
	}{
		{"movies by rating", ListFilter{Type: "movie", Sort: "rating"}, []string{"Alien", "Heat"}},
		{"planned", ListFilter{Status: "planned"}, []string{"Dune"}},
		{"search creator", ListFilter{Query: "mann"}, []string{"Heat"}},
		{"favorites", ListFilter{Favorite: ptr(true)}, []string{"Heat"}},
		{"title sort", ListFilter{Sort: "title"}, []string{"Alien", "Dune", "Heat", "The Wire"}},
	}
	for _, tc := range cases {
		items, err := ListMedia(ctxFor(db, 1), tc.filter)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		got := titles(items)
		if len(got) != len(tc.want) {
			t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
			}
		}
	}
	if _, err := ListMedia(ctxFor(db, 1), ListFilter{Sort: "bogus"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid for bad sort, got %v", err)
	}

	stats, err := GetMediaStats(ctxFor(db, 1), now)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != 4 || stats.ByType["movie"] != 2 || stats.ByStatus["completed"] != 2 || stats.CompletedThisYear != 2 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if stats.AverageRating == nil || *stats.AverageRating != 4.75 {
		t.Fatalf("unexpected average rating: %v", stats.AverageRating)
	}
}

func TestDeleteMissing(t *testing.T) {
	db := setupTestDB(t)
	if err := DeleteMedia(ctxFor(db, 1), 99); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func ptr[T any](v T) *T { return &v }
