package steps

import (
	"context"
	"net/http"
	"testing"

	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Two-user tests (Postgres only, TEST_DATABASE_URL): Alice's data, Bob's token.

func stepsAPI(t *testing.T) (*twousers.Harness, *twousers.API) {
	h := twousers.New(t)
	api := h.Router(func(g *gin.RouterGroup) { RegisterStepsRoutes(g.Group("/steps")) })
	return h, api
}

type stepsJSON struct {
	Steps  int   `json:"steps"`
	UserID *uint `json:"user_id"`
}

func listSteps(t *testing.T, api *twousers.API, u twousers.User) []stepsJSON {
	t.Helper()
	var out struct{ Logs []stepsJSON }
	api.Do(u, http.MethodGet, "/steps", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Logs
}

func TestSteps_eachUserSeesOnlyTheirOwn(t *testing.T) {
	h, api := stepsAPI(t)
	api.Do(h.Alice, http.MethodPost, "/steps", map[string]any{"date": "2026-01-15", "steps": 9000}).Status(t, http.StatusOK)
	if got := listSteps(t, api, h.Bob); len(got) != 0 {
		t.Fatalf("Bob sees %+v, want none", got)
	}
	// Same day for Bob is his own row.
	api.Do(h.Bob, http.MethodPost, "/steps", map[string]any{"date": "2026-01-15", "steps": 4000}).Status(t, http.StatusOK)
	for _, tc := range []struct {
		u    twousers.User
		want int
	}{{h.Alice, 9000}, {h.Bob, 4000}} {
		got := listSteps(t, api, tc.u)
		if len(got) != 1 || got[0].Steps != tc.want || got[0].UserID != nil {
			t.Fatalf("%s sees %+v, want one log of %d and no user_id", tc.u.Username, got, tc.want)
		}
	}
	api.Do(h.Bob, http.MethodPost, "/steps", map[string]any{"date": "2026-01-15", "steps": 4100}).Status(t, http.StatusOK)
	if a := listSteps(t, api, h.Alice); a[0].Steps != 9000 {
		t.Fatalf("Bob's update changed Alice's log: %+v", a)
	}
	api.Do(twousers.Anonymous, http.MethodGet, "/steps", nil).Status(t, http.StatusUnauthorized)
}

func TestSteps_createIgnoresUserIDInBody(t *testing.T) {
	h, api := stepsAPI(t)
	api.Do(h.Bob, http.MethodPost, "/steps", map[string]any{"date": "2026-01-15", "steps": 10, "user_id": h.Alice.ID}).
		Status(t, http.StatusOK)
	var owner uint
	if err := h.Owner.Raw("SELECT user_id FROM step_logs").Scan(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if owner != h.Bob.ID {
		t.Fatalf("log owned by %d, want Bob (%d)", owner, h.Bob.ID)
	}
}

func TestSteps_rlsBackstop(t *testing.T) {
	h, api := stepsAPI(t)
	api.Do(h.Alice, http.MethodPost, "/steps", map[string]any{"date": "2026-01-15", "steps": 9000}).Status(t, http.StatusOK)
	err := h.Run(h.Bob, func(ctx context.Context) error {
		var rows []StepLog
		if err := database.From(ctx).Unscoped().Find(&rows).Error; err != nil { // no Owned scope
			return err
		}
		if len(rows) != 0 {
			t.Errorf("unscoped query as Bob returned %d of Alice's rows", len(rows))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec("INSERT INTO step_logs (date, steps, user_id) VALUES (now(), 1, ?)", h.Alice.ID).Error
	})
	if err == nil {
		t.Fatal("insert of a row for Alice as Bob succeeded; want a row-level security error")
	}
}
