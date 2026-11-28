package weight

import (
	"context"
	"net/http"
	"testing"

	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Two-user tests (Postgres only, TEST_DATABASE_URL): Alice's data, Bob's token.

func weightAPI(t *testing.T) (*twousers.Harness, *twousers.API) {
	h := twousers.New(t)
	api := h.Router(func(g *gin.RouterGroup) { RegisterWeightRoutes(g.Group("/weight")) })
	return h, api
}

type weightJSON struct {
	WeightLbs  float64  `json:"weight_lbs"`
	Date       string   `json:"date"`
	BodyFatPct *float64 `json:"body_fat_pct"`
	UserID     *uint    `json:"user_id"`
}

func listWeights(t *testing.T, api *twousers.API, u twousers.User) []weightJSON {
	t.Helper()
	var out struct{ Logs []weightJSON }
	api.Do(u, http.MethodGet, "/weight", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Logs
}

func TestWeight_eachUserSeesOnlyTheirOwn(t *testing.T) {
	h, api := weightAPI(t)
	api.Do(h.Alice, http.MethodPost, "/weight", map[string]any{"date": "2026-01-15", "weight_lbs": 180}).Status(t, http.StatusOK)
	if got := listWeights(t, api, h.Bob); len(got) != 0 {
		t.Fatalf("Bob sees %+v, want none", got)
	}
	// The same day for Bob is his own row, not a conflict with Alice's.
	api.Do(h.Bob, http.MethodPost, "/weight", map[string]any{"date": "2026-01-15", "weight_lbs": 150}).Status(t, http.StatusOK)
	for _, tc := range []struct {
		u    twousers.User
		want float64
	}{{h.Alice, 180}, {h.Bob, 150}} {
		got := listWeights(t, api, tc.u)
		if len(got) != 1 || got[0].WeightLbs != tc.want || got[0].UserID != nil {
			t.Fatalf("%s sees %+v, want one log of %v and no user_id", tc.u.Username, got, tc.want)
		}
	}
	if got := listWeights(t, api, h.Admin); len(got) != 0 {
		t.Fatalf("admin sees %+v, want none", got)
	}
	api.Do(twousers.Anonymous, http.MethodGet, "/weight", nil).Status(t, http.StatusUnauthorized)

	// Updating your own day doesn't touch anyone else's.
	api.Do(h.Bob, http.MethodPost, "/weight", map[string]any{"date": "2026-01-15", "weight_lbs": 151}).Status(t, http.StatusOK)
	if a, b := listWeights(t, api, h.Alice), listWeights(t, api, h.Bob); a[0].WeightLbs != 180 || len(b) != 1 || b[0].WeightLbs != 151 {
		t.Fatalf("after Bob's update: Alice %+v, Bob %+v", a, b)
	}
}

func TestWeight_createIgnoresUserIDInBody(t *testing.T) {
	h, api := weightAPI(t)
	api.Do(h.Bob, http.MethodPost, "/weight", map[string]any{"date": "2026-01-15", "weight_lbs": 150, "user_id": h.Alice.ID}).
		Status(t, http.StatusOK)
	var owner uint
	if err := h.Owner.Raw("SELECT user_id FROM body_weight_logs").Scan(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if owner != h.Bob.ID {
		t.Fatalf("log owned by %d, want Bob (%d)", owner, h.Bob.ID)
	}
	if got := listWeights(t, api, h.Alice); len(got) != 0 {
		t.Fatalf("Alice sees %+v", got)
	}
}

// The import writes as the importing user, whatever the same days hold for someone else.
func TestWeight_importWritesAsTheImportingUser(t *testing.T) {
	h, api := weightAPI(t)
	api.Do(h.Alice, http.MethodPost, "/weight", map[string]any{"date": "2026-11-19", "weight_lbs": 190}).Status(t, http.StatusOK)

	imp := map[string]any{"csv": renphoHeader + renphoRow}
	var out struct{ Created, Updated int }
	api.Do(h.Bob, http.MethodPost, "/weight/import", imp).Status(t, http.StatusOK).JSON(t, &out)
	if out.Created != 1 || out.Updated != 0 {
		t.Fatalf("Bob's import: %+v, want 1 created (Alice's same day is not his)", out)
	}
	// Re-import updates Bob's row only.
	api.Do(h.Bob, http.MethodPost, "/weight/import", map[string]any{"csv": "Date,Weight(lb)\n2026.11.19,163.0\n"}).
		Status(t, http.StatusOK).JSON(t, &out)
	if out.Created != 0 || out.Updated != 1 {
		t.Fatalf("Bob's re-import: %+v, want 1 updated", out)
	}

	a, b := listWeights(t, api, h.Alice), listWeights(t, api, h.Bob)
	if len(a) != 1 || a[0].WeightLbs != 190 || a[0].BodyFatPct != nil {
		t.Fatalf("Alice's log changed by Bob's import: %+v", a)
	}
	if len(b) != 1 || b[0].WeightLbs != 163 || b[0].BodyFatPct != nil {
		t.Fatalf("Bob's log: %+v", b)
	}
	var owners []uint
	if err := h.Owner.Raw("SELECT user_id FROM body_weight_logs ORDER BY id").Scan(&owners).Error; err != nil {
		t.Fatal(err)
	}
	if len(owners) != 2 || owners[0] != h.Alice.ID || owners[1] != h.Bob.ID {
		t.Fatalf("owners = %v, want [Alice Bob]", owners)
	}
}

func TestWeight_rlsBackstop(t *testing.T) {
	h, api := weightAPI(t)
	api.Do(h.Alice, http.MethodPost, "/weight", map[string]any{"date": "2026-01-15", "weight_lbs": 180}).Status(t, http.StatusOK)

	err := h.Run(h.Bob, func(ctx context.Context) error {
		var rows []BodyWeightLog
		if err := database.From(ctx).Unscoped().Find(&rows).Error; err != nil { // no Owned scope
			return err
		}
		if len(rows) != 0 {
			t.Errorf("unscoped query as Bob returned %d of Alice's rows", len(rows))
		}
		if n := database.From(ctx).Exec("UPDATE body_weight_logs SET weight_lbs = 1").RowsAffected; n != 0 {
			t.Errorf("unscoped update as Bob touched %d rows", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec(
			"INSERT INTO body_weight_logs (date, weight_lbs, user_id) VALUES (now(), 1, ?)", h.Alice.ID).Error
	})
	if err == nil {
		t.Fatal("insert of a row for Alice as Bob succeeded; want a row-level security error")
	}
	var n int64
	if err := h.App.Model(&BodyWeightLog{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("app role without app.user_id sees %d logs", n)
	}
}
