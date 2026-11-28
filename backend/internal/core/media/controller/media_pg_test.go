package controller

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"be-simpletracker/internal/core/media/lookup"
	"be-simpletracker/internal/core/media/models"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Two-user tests for media (docs/multi-user/DECISIONS.md "C5 media"): a shared catalog and each
// user's own saves of it. Postgres only (TEST_DATABASE_URL).

func mediaAPI(t *testing.T) (*twousers.Harness, *twousers.API) {
	h := twousers.New(t)
	api := h.Router(func(g *gin.RouterGroup) { RegisterMediaRoutes(g.Group("/media"), lookup.NewService("")) })
	return h, api
}

type entryJSON struct {
	ID          uint     `json:"ID"`
	MediaItemID uint     `json:"media_item_id"`
	CreatedBy   uint     `json:"created_by"`
	CanEdit     bool     `json:"can_edit_catalog"`
	Title       string   `json:"title"`
	CoverURL    string   `json:"cover_url"`
	Status      string   `json:"status"`
	Rating      *float64 `json:"rating"`
	Review      string   `json:"review"`
	UserID      *uint    `json:"user_id"`
}

type catalogJSON struct {
	ID      uint   `json:"ID"`
	Title   string `json:"title"`
	SavedID *uint  `json:"saved_id"`
	CanEdit bool   `json:"can_edit_catalog"`
}

func save(t *testing.T, api *twousers.API, u twousers.User, body map[string]any, want int) entryJSON {
	t.Helper()
	var out struct{ Item entryJSON }
	api.Do(u, http.MethodPost, "/media", body).Status(t, want).JSON(t, &out)
	return out.Item
}

func list(t *testing.T, api *twousers.API, u twousers.User) []entryJSON {
	t.Helper()
	var out struct{ Items []entryJSON }
	api.Do(u, http.MethodGet, "/media", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Items
}

func catalog(t *testing.T, api *twousers.API, u twousers.User, q string) []catalogJSON {
	t.Helper()
	var out struct{ Items []catalogJSON }
	api.Do(u, http.MethodGet, "/media/catalog?q="+q, nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Items
}

var heat = map[string]any{"type": "movie", "title": "Heat", "source": "tmdb", "external_id": "movie/949", "status": "completed", "rating": 4.5, "review": "Alice's private take"}

func TestMedia_eachUserSeesOnlyTheirOwnSaves(t *testing.T) {
	h, api := mediaAPI(t)
	save(t, api, h.Alice, heat, http.StatusCreated)

	if got := list(t, api, h.Alice); len(got) != 1 || got[0].Status != "completed" || got[0].UserID != nil {
		t.Fatalf("Alice sees %+v", got)
	}
	if got := list(t, api, h.Bob); len(got) != 0 {
		t.Fatalf("Bob sees %+v, want none", got)
	}
	var stats struct{ Stats struct{ Total int64 } }
	api.Do(h.Bob, http.MethodGet, "/media/stats", nil).Status(t, http.StatusOK).JSON(t, &stats)
	if stats.Stats.Total != 0 {
		t.Fatalf("Bob's stats count %d", stats.Stats.Total)
	}
	api.Do(twousers.Anonymous, http.MethodGet, "/media", nil).Status(t, http.StatusUnauthorized)
	api.Do(twousers.Anonymous, http.MethodGet, "/media/catalog", nil).Status(t, http.StatusUnauthorized)
}

func TestMedia_someoneElsesSaveIsNotFound(t *testing.T) {
	h, api := mediaAPI(t)
	alices := save(t, api, h.Alice, heat, http.StatusCreated)
	missing := alices.ID + 1000

	for _, id := range []uint{alices.ID, missing} {
		path := fmt.Sprintf("/media/%d", id)
		api.Do(h.Bob, http.MethodGet, path, nil).NotFound(t)
		api.Do(h.Bob, http.MethodPatch, path, map[string]any{"rating": 1}).NotFound(t)
		api.Do(h.Bob, http.MethodDelete, path, nil).NotFound(t)
	}
	got := list(t, api, h.Alice)
	if len(got) != 1 || got[0].Rating == nil || *got[0].Rating != 4.5 || got[0].Review != "Alice's private take" {
		t.Fatalf("Alice's save changed: %+v", got)
	}
	path := fmt.Sprintf("/media/%d", alices.ID)
	api.Do(h.Alice, http.MethodPatch, path, map[string]any{"rating": 5}).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, path, nil).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, path, nil).NotFound(t)
}

// Both can save the same catalog item independently; the catalog row is shared.
func TestMedia_sameCatalogItemSavedIndependently(t *testing.T) {
	h, api := mediaAPI(t)
	a := save(t, api, h.Alice, heat, http.StatusCreated)
	b := save(t, api, h.Bob, map[string]any{"type": "movie", "title": "Heat (dupe)", "source": "tmdb", "external_id": "movie/949", "status": "dropped"}, http.StatusCreated)
	if a.MediaItemID != b.MediaItemID {
		t.Fatalf("same external ID made two catalog rows (%d, %d)", a.MediaItemID, b.MediaItemID)
	}
	if b.Title != "Heat" || b.Rating != nil || b.Review != "" {
		t.Fatalf("Bob's save leaked Alice's data or overwrote the catalog: %+v", b)
	}
	// Saving the same thing again returns the existing save.
	again := save(t, api, h.Bob, map[string]any{"type": "movie", "title": "Heat", "source": "tmdb", "external_id": "movie/949"}, http.StatusOK)
	if again.ID != b.ID || again.Status != "dropped" {
		t.Fatalf("second save: %+v, want Bob's existing %d", again, b.ID)
	}
	api.Do(h.Bob, http.MethodPatch, fmt.Sprintf("/media/%d", b.ID), map[string]any{"status": "in_progress", "rating": 2}).Status(t, http.StatusOK)
	if got := list(t, api, h.Alice); len(got) != 1 || got[0].Status != "completed" || *got[0].Rating != 4.5 {
		t.Fatalf("Bob's edit changed Alice's save: %+v", got)
	}
	var n int64
	h.Owner.Table("media_items").Count(&n)
	if n != 1 {
		t.Fatalf("%d catalog rows, want 1", n)
	}
}

func TestMedia_catalogSearchAndSaveExisting(t *testing.T) {
	h, api := mediaAPI(t)
	a := save(t, api, h.Alice, heat, http.StatusCreated)

	got := catalog(t, api, h.Bob, "hea")
	if len(got) != 1 || got[0].ID != a.MediaItemID || got[0].SavedID != nil || got[0].CanEdit {
		t.Fatalf("Bob's catalog search: %+v", got)
	}
	if mine := catalog(t, api, h.Alice, "hea"); len(mine) != 1 || mine[0].SavedID == nil || *mine[0].SavedID != a.ID || !mine[0].CanEdit {
		t.Fatalf("Alice's catalog search: %+v", mine)
	}
	var out struct{ Item entryJSON }
	api.Do(h.Bob, http.MethodPost, fmt.Sprintf("/media/catalog/%d/save", a.MediaItemID), nil).Status(t, http.StatusCreated).JSON(t, &out)
	if out.Item.Status != "planned" || out.Item.Rating != nil || out.Item.Review != "" {
		t.Fatalf("Bob's fresh save: %+v", out.Item)
	}
	api.Do(h.Bob, http.MethodPost, fmt.Sprintf("/media/catalog/%d/save", a.MediaItemID+1000), nil).NotFound(t)
	if got := catalog(t, api, h.Bob, "hea"); got[0].SavedID == nil || *got[0].SavedID != out.Item.ID {
		t.Fatalf("saved_id not shown after saving: %+v", got)
	}
}

func TestMedia_catalogEditAndDeletePermissions(t *testing.T) {
	h, api := mediaAPI(t)
	a := save(t, api, h.Alice, heat, http.StatusCreated)
	path := fmt.Sprintf("/media/catalog/%d", a.MediaItemID)

	// Bob can't edit or delete Alice's entry; a missing one is a 404.
	api.Do(h.Bob, http.MethodPatch, path, map[string]any{"title": "Bob's Heat"}).Status(t, http.StatusForbidden)
	api.Do(h.Bob, http.MethodDelete, path, nil).Status(t, http.StatusForbidden)
	api.Do(h.Bob, http.MethodPatch, fmt.Sprintf("/media/catalog/%d", a.MediaItemID+1000), map[string]any{"title": "x"}).NotFound(t)
	// The creator edits but can't delete.
	api.Do(h.Alice, http.MethodPatch, path, map[string]any{"title": "Heat (1995)"}).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, path, nil).Status(t, http.StatusForbidden)
	// An admin edits anyone's, and is the only one who deletes.
	api.Do(h.Admin, http.MethodPatch, path, map[string]any{"genres": "Crime"}).Status(t, http.StatusOK)
	if got := list(t, api, h.Alice); got[0].Title != "Heat (1995)" {
		t.Fatalf("title not updated: %+v", got)
	}
	api.Do(h.Admin, http.MethodDelete, path, nil).Status(t, http.StatusOK)
	if got := list(t, api, h.Alice); len(got) != 0 {
		t.Fatalf("deleted catalog entry still in Alice's list: %+v", got)
	}
	if got := catalog(t, api, h.Bob, "heat"); len(got) != 0 {
		t.Fatalf("deleted catalog entry still searchable: %+v", got)
	}
	api.Do(h.Admin, http.MethodDelete, path, nil).NotFound(t)
}

func TestMedia_forgedOwnerFieldsAreIgnored(t *testing.T) {
	h, api := mediaAPI(t)
	body := map[string]any{"type": "book", "title": "Dune", "user_id": h.Alice.ID, "created_by": h.Alice.ID}
	b := save(t, api, h.Bob, body, http.StatusCreated)
	var userID, createdBy uint
	h.Owner.Raw("SELECT user_id FROM user_media WHERE id = ?", b.ID).Scan(&userID)
	h.Owner.Raw("SELECT created_by FROM media_items WHERE id = ?", b.MediaItemID).Scan(&createdBy)
	if userID != h.Bob.ID || createdBy != h.Bob.ID {
		t.Fatalf("user_id %d, created_by %d: want Bob (%d)", userID, createdBy, h.Bob.ID)
	}
	// Nor can a catalog edit hand the entry over.
	api.Do(h.Bob, http.MethodPatch, fmt.Sprintf("/media/catalog/%d", b.MediaItemID), map[string]any{"created_by": h.Alice.ID, "title": "Dune 2"}).Status(t, http.StatusOK)
	h.Owner.Raw("SELECT created_by FROM media_items WHERE id = ?", b.MediaItemID).Scan(&createdBy)
	if createdBy != h.Bob.ID {
		t.Fatalf("created_by changed to %d", createdBy)
	}
	if got := list(t, api, h.Alice); len(got) != 0 {
		t.Fatalf("Alice sees %+v", got)
	}
}

// BACKLOG #15: a catalog cover is shown to everyone, so only lookup-source image hosts are accepted.
func TestMedia_coverURLsAreValidated(t *testing.T) {
	h, api := mediaAPI(t)
	for _, bad := range []string{"https://evil.example/pixel.gif", "http://image.tmdb.org/t/p/w500/a.jpg", "javascript:alert(1)"} {
		api.Do(h.Alice, http.MethodPost, "/media", map[string]any{"type": "movie", "title": "x", "cover_url": bad}).Status(t, http.StatusBadRequest)
	}
	good := "https://image.tmdb.org/t/p/w500/a.jpg"
	a := save(t, api, h.Alice, map[string]any{"type": "movie", "title": "x", "cover_url": good}, http.StatusCreated)
	path := fmt.Sprintf("/media/catalog/%d", a.MediaItemID)
	api.Do(h.Alice, http.MethodPatch, path, map[string]any{"cover_url": "https://evil.example/p.gif"}).Status(t, http.StatusBadRequest)
	api.Do(h.Alice, http.MethodPatch, path, map[string]any{"cover_url": nil}).Status(t, http.StatusOK)

	// A stored cover from before the rule is never handed out.
	if err := h.Owner.Exec("UPDATE media_items SET cover_url = 'https://evil.example/old.gif' WHERE id = ?", a.MediaItemID).Error; err != nil {
		t.Fatal(err)
	}
	if got := list(t, api, h.Bob); len(got) != 0 {
		t.Fatal("Bob has no saves")
	}
	api.Do(h.Bob, http.MethodPost, fmt.Sprintf("/media/catalog/%d/save", a.MediaItemID), nil).Status(t, http.StatusCreated)
	if got := list(t, api, h.Bob); got[0].CoverURL != "" {
		t.Fatalf("legacy cover leaked: %q", got[0].CoverURL)
	}
	if got := catalog(t, api, h.Bob, "x"); len(got) != 1 {
		t.Fatalf("catalog: %+v", got)
	}
}

// Row-level security is the backstop: unscoped queries as Bob see none of Alice's saves, the
// catalog is readable, and he can't write a row for Alice or change her entry.
func TestMedia_rlsBackstop(t *testing.T) {
	h, api := mediaAPI(t)
	a := save(t, api, h.Alice, heat, http.StatusCreated)

	err := h.Run(h.Bob, func(ctx context.Context) error {
		var saves []models.UserMedia
		if err := database.From(ctx).Unscoped().Find(&saves).Error; err != nil { // no Owned scope
			return err
		}
		if len(saves) != 0 {
			t.Errorf("unscoped query as Bob returned %d of Alice's saves", len(saves))
		}
		if n := database.From(ctx).Exec("UPDATE user_media SET rating = 1").RowsAffected; n != 0 {
			t.Errorf("unscoped update as Bob touched %d saves", n)
		}
		var items []models.MediaItem
		if err := database.From(ctx).Find(&items).Error; err != nil || len(items) != 1 {
			t.Errorf("Bob reads the catalog: %d rows, %v", len(items), err)
		}
		if n := database.From(ctx).Exec("UPDATE media_items SET title = 'x'").RowsAffected; n != 0 {
			t.Errorf("Bob edited %d of Alice's catalog rows", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec(
			"INSERT INTO user_media (user_id, media_item_id, status) VALUES (?, ?, 'planned')", h.Alice.ID, a.MediaItemID).Error
	}); err == nil {
		t.Fatal("inserting a save for Alice as Bob succeeded; want a row-level security error")
	}
	if err := h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec(
			"INSERT INTO media_items (type, title, created_by) VALUES ('movie', 'x', ?)", h.Alice.ID).Error
	}); err == nil {
		t.Fatal("adding a catalog row as Alice while being Bob succeeded")
	}
	// Soft delete is an UPDATE: catalog_guard refuses it for a non-admin.
	if err := h.Run(h.Alice, func(ctx context.Context) error {
		return database.From(ctx).Exec("UPDATE media_items SET deleted_at = now()").Error
	}); err == nil {
		t.Fatal("a non-admin soft-deleted a catalog row")
	}
	var n int64
	if err := h.App.Model(&models.UserMedia{}).Count(&n).Error; err != nil || n != 0 {
		t.Fatalf("app role without app.user_id sees %d saves (%v)", n, err)
	}
}

// A row from before the catalog split can have half a (source, external_id) pair; editing it
// must keep working, and an edit must bump updated_at and write NULLs (cleared rating).
func TestMedia_legacyHalfPairStaysEditable(t *testing.T) {
	h, api := mediaAPI(t)
	a := save(t, api, h.Alice, map[string]any{"type": "movie", "title": "Old", "rating": 3}, http.StatusCreated)
	if err := h.Owner.Exec("UPDATE media_items SET source = 'tmdb' WHERE id = ?", a.MediaItemID).Error; err != nil {
		t.Fatal(err)
	}
	path := fmt.Sprintf("/media/catalog/%d", a.MediaItemID)
	api.Do(h.Alice, http.MethodPatch, path, map[string]any{"title": "Older"}).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodPatch, path, map[string]any{"external_id": ""}).Status(t, http.StatusBadRequest)
	api.Do(h.Alice, http.MethodPatch, fmt.Sprintf("/media/%d", a.ID), map[string]any{"rating": nil}).Status(t, http.StatusOK)
	got := list(t, api, h.Alice)
	if len(got) != 1 || got[0].Title != "Older" || got[0].Rating != nil {
		t.Fatalf("after edits: %+v", got)
	}
	var bumped bool
	h.Owner.Raw("SELECT updated_at > created_at FROM user_media WHERE id = ?", a.ID).Scan(&bumped)
	if !bumped {
		t.Fatal("updated_at not bumped by a personal edit")
	}
}
