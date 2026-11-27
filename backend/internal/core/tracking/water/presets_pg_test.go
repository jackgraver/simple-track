package water

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Two-user tests for drink size presets: the reference for every owned endpoint
// (docs/multi-user/DECISIONS.md, "B3 owner kit"). Postgres only (TEST_DATABASE_URL).

func presetAPI(t *testing.T) (*twousers.Harness, *twousers.API) {
	h := twousers.New(t)
	api := h.Router(func(g *gin.RouterGroup) { RegisterWaterRoutes(g.Group("/water")) })
	return h, api
}

type presetJSON struct {
	ID       uint    `json:"ID"`
	Name     string  `json:"name"`
	AmountOz float64 `json:"amount_oz"`
	UserID   *uint   `json:"user_id"`
}

func createPreset(t *testing.T, api *twousers.API, u twousers.User, name string, oz float64) presetJSON {
	t.Helper()
	var out struct{ Preset presetJSON }
	api.Do(u, http.MethodPost, "/water/presets", map[string]any{"name": name, "amount_oz": oz}).
		Status(t, http.StatusOK).JSON(t, &out)
	return out.Preset
}

func listPresets(t *testing.T, api *twousers.API, u twousers.User) []presetJSON {
	t.Helper()
	var out struct{ Presets []presetJSON }
	api.Do(u, http.MethodGet, "/water/presets", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Presets
}

func TestPresets_eachUserSeesOnlyTheirOwn(t *testing.T) {
	h, api := presetAPI(t)
	createPreset(t, api, h.Alice, "Alice bottle", 24)
	createPreset(t, api, h.Bob, "Bob cup", 8)

	for _, tc := range []struct {
		u    twousers.User
		want string
	}{{h.Alice, "Alice bottle"}, {h.Bob, "Bob cup"}} {
		got := listPresets(t, api, tc.u)
		if len(got) != 1 || got[0].Name != tc.want {
			t.Fatalf("%s sees %+v, want only %q", tc.u.Username, got, tc.want)
		}
		if got[0].UserID != nil {
			t.Fatalf("user_id is in the response: %+v", got[0])
		}
	}
	if got := listPresets(t, api, h.Admin); len(got) != 0 {
		t.Fatalf("admin sees %+v, want none", got)
	}
	api.Do(twousers.Anonymous, http.MethodGet, "/water/presets", nil).Status(t, http.StatusUnauthorized)
}

func TestPresets_someoneElsesIsNotFound(t *testing.T) {
	h, api := presetAPI(t)
	alices := createPreset(t, api, h.Alice, "Bottle", 24)
	missing := alices.ID + 1000

	for _, id := range []uint{alices.ID, missing} {
		path := fmt.Sprintf("/water/presets/%d", id)
		bob := api.Do(h.Bob, http.MethodPut, path, map[string]any{"name": "Mine now", "amount_oz": 1})
		bob.NotFound(t)
		api.Do(h.Bob, http.MethodDelete, path, nil).NotFound(t)
	}

	// Nothing changed for Alice, and her own update and delete still work.
	got := listPresets(t, api, h.Alice)
	if len(got) != 1 || got[0].Name != "Bottle" || got[0].AmountOz != 24 {
		t.Fatalf("Alice's preset changed: %+v", got)
	}
	path := fmt.Sprintf("/water/presets/%d", alices.ID)
	api.Do(h.Alice, http.MethodPut, path, map[string]any{"name": "Big bottle", "amount_oz": 32}).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, path, nil).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, path, nil).NotFound(t)
}

func TestPresets_bobCantLogWaterWithAlicesPreset(t *testing.T) {
	h, api := presetAPI(t)
	alices := createPreset(t, api, h.Alice, "Bottle", 24)
	body := map[string]any{"date": "2026-01-15", "amount_oz": 24, "preset_id": alices.ID}
	res := api.Do(h.Bob, http.MethodPost, "/water", body).Status(t, http.StatusBadRequest)
	missing := api.Do(h.Bob, http.MethodPost, "/water", map[string]any{"date": "2026-01-15", "amount_oz": 24, "preset_id": alices.ID + 1000}).
		Status(t, http.StatusBadRequest)
	if string(res.Body) != string(missing.Body) {
		t.Fatalf("someone else's preset answers %s, a missing one %s: must be the same", res.Body, missing.Body)
	}
	api.Do(h.Alice, http.MethodPost, "/water", body).Status(t, http.StatusOK)
}

// A forged user_id in the body is ignored: the row belongs to whoever is signed in.
func TestPresets_createIgnoresUserIDInBody(t *testing.T) {
	h, api := presetAPI(t)
	var out struct{ Preset presetJSON }
	api.Do(h.Bob, http.MethodPost, "/water/presets", map[string]any{"name": "Sneaky", "amount_oz": 8, "user_id": h.Alice.ID}).
		Status(t, http.StatusOK).JSON(t, &out)
	var owner uint
	if err := h.Owner.Raw("SELECT user_id FROM drink_size_presets WHERE id = ?", out.Preset.ID).Scan(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if owner != h.Bob.ID {
		t.Fatalf("preset owned by %d, want Bob (%d)", owner, h.Bob.ID)
	}
	if got := listPresets(t, api, h.Alice); len(got) != 0 {
		t.Fatalf("Alice sees %+v", got)
	}
}

// Row-level security is the backstop: a query that forgets database.Owned still sees only the
// caller's rows, and an insert for someone else is refused by the database.
func TestPresets_rlsBackstop(t *testing.T) {
	h, api := presetAPI(t)
	createPreset(t, api, h.Alice, "Bottle", 24)

	err := h.Run(h.Bob, func(ctx context.Context) error {
		var rows []DrinkSizePreset
		if err := database.From(ctx).Unscoped().Find(&rows).Error; err != nil { // no Owned scope
			return err
		}
		if len(rows) != 0 {
			t.Errorf("unscoped query as Bob returned %d of Alice's rows", len(rows))
		}
		if n := database.From(ctx).Exec("UPDATE drink_size_presets SET name = 'x'").RowsAffected; n != 0 {
			t.Errorf("unscoped update as Bob touched %d rows", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	err = h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec(
			"INSERT INTO drink_size_presets (name, amount_oz, user_id) VALUES ('x', 1, ?)", h.Alice.ID).Error
	})
	if err == nil {
		t.Fatal("insert of a row for Alice as Bob succeeded; want a row-level security error")
	}

	// Without an identity the app role sees nothing at all.
	var n int64
	if err := h.App.Model(&DrinkSizePreset{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("app role without app.user_id sees %d presets", n)
	}
}
