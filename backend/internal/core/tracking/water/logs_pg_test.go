package water

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"
)

// Two-user tests for water logs (Postgres only, TEST_DATABASE_URL). Presets: presets_pg_test.go.

type logJSON struct {
	ID       uint        `json:"ID"`
	AmountOz float64     `json:"amount_oz"`
	ClientID string      `json:"client_id"`
	Preset   *presetJSON `json:"preset"`
	UserID   *uint       `json:"user_id"`
}

const day = "2026-01-15"

func postLog(t *testing.T, api *twousers.API, u twousers.User, body map[string]any) logJSON {
	t.Helper()
	var out struct{ Log logJSON }
	api.Do(u, http.MethodPost, "/water", body).Status(t, http.StatusOK).JSON(t, &out)
	return out.Log
}

func listLogs(t *testing.T, api *twousers.API, u twousers.User) []logJSON {
	t.Helper()
	var out struct{ Logs []logJSON }
	api.Do(u, http.MethodGet, "/water?date="+day, nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Logs
}

func TestWaterLogs_eachUserSeesOnlyTheirOwn(t *testing.T) {
	h, api := presetAPI(t)
	bottle := createPreset(t, api, h.Alice, "Bottle", 24)
	postLog(t, api, h.Alice, map[string]any{"date": day, "amount_oz": 24, "preset_id": bottle.ID})
	if got := listLogs(t, api, h.Bob); len(got) != 0 {
		t.Fatalf("Bob sees %+v, want none", got)
	}
	postLog(t, api, h.Bob, map[string]any{"date": day, "amount_oz": 8})
	for _, tc := range []struct {
		u  twousers.User
		oz float64
	}{{h.Alice, 24}, {h.Bob, 8}} {
		got := listLogs(t, api, tc.u)
		if len(got) != 1 || got[0].AmountOz != tc.oz || got[0].UserID != nil {
			t.Fatalf("%s sees %+v, want one log of %v and no user_id", tc.u.Username, got, tc.oz)
		}
	}
	if got := listLogs(t, api, h.Alice); got[0].Preset == nil || got[0].Preset.Name != "Bottle" {
		t.Fatalf("Alice's log lost its preset: %+v", got[0])
	}
	api.Do(twousers.Anonymous, http.MethodGet, "/water?date="+day, nil).Status(t, http.StatusUnauthorized)
}

func TestWaterLogs_someoneElsesIsNotFound(t *testing.T) {
	h, api := presetAPI(t)
	alices := postLog(t, api, h.Alice, map[string]any{"date": day, "amount_oz": 16, "client_id": "6f1c2c1e-0000-4000-8000-0000000000a1"})
	missing := alices.ID + 1000

	byID := func(id uint) string { return fmt.Sprintf("/water/%d", id) }
	// NotFound checks the exact body, so a missing log and Alice's answer the same.
	api.Do(h.Bob, http.MethodDelete, byID(missing), nil).NotFound(t)
	api.Do(h.Bob, http.MethodDelete, byID(alices.ID), nil).NotFound(t)
	// By client_id, too.
	api.Do(h.Bob, http.MethodDelete, "/water/"+alices.ClientID, nil).NotFound(t)
	api.Do(h.Bob, http.MethodDelete, "/water/6f1c2c1e-0000-4000-8000-0000000000ff", nil).NotFound(t)

	if got := listLogs(t, api, h.Alice); len(got) != 1 {
		t.Fatalf("Alice's log was deleted by Bob: %+v", got)
	}
	api.Do(h.Alice, http.MethodDelete, byID(alices.ID), nil).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodDelete, byID(alices.ID), nil).NotFound(t)
}

// The same client_id is two different rows for two users: Bob's create doesn't return (or collide
// with) Alice's, and each can delete theirs by it.
func TestWaterLogs_sameClientIDForTwoUsers(t *testing.T) {
	h, api := presetAPI(t)
	const cid = "6f1c2c1e-0000-4000-8000-0000000000b2"
	a := postLog(t, api, h.Alice, map[string]any{"date": day, "amount_oz": 16, "client_id": cid})
	b := postLog(t, api, h.Bob, map[string]any{"date": day, "amount_oz": 8, "client_id": cid})
	if a.ID == b.ID || b.AmountOz != 8 {
		t.Fatalf("Bob's create with Alice's client_id returned %+v (Alice's is %+v)", b, a)
	}
	// Each user's replay is still idempotent.
	if again := postLog(t, api, h.Bob, map[string]any{"date": day, "amount_oz": 8, "client_id": cid}); again.ID != b.ID {
		t.Fatalf("Bob's replay created a new row: %d vs %d", again.ID, b.ID)
	}
	if got := listLogs(t, api, h.Alice); len(got) != 1 || got[0].ID != a.ID {
		t.Fatalf("Alice sees %+v", got)
	}
	api.Do(h.Bob, http.MethodDelete, "/water/"+cid, nil).Status(t, http.StatusOK)
	if got := listLogs(t, api, h.Alice); len(got) != 1 {
		t.Fatalf("Bob's delete by client_id removed Alice's log: %+v", got)
	}
}

// Bob can't point a log at Alice's preset, in the API (400 like a missing one: presets_pg_test)
// or in the database (composite FK), even with a forged user_id.
func TestWaterLogs_cantAttachToSomeoneElsesPreset(t *testing.T) {
	h, api := presetAPI(t)
	alices := createPreset(t, api, h.Alice, "Bottle", 24)
	api.Do(h.Bob, http.MethodPost, "/water", map[string]any{"date": day, "amount_oz": 8, "preset_id": alices.ID, "user_id": h.Alice.ID}).
		Status(t, http.StatusBadRequest)

	err := h.Owner.Exec(`INSERT INTO water_logs (date, amount_oz, preset_id, user_id) VALUES (now(), 8, ?, ?)`,
		alices.ID, h.Bob.ID).Error
	if err == nil {
		t.Fatal("a log for Bob with Alice's preset was accepted; want the composite foreign key to refuse it")
	}
}

func TestWaterLogs_createIgnoresUserIDInBody(t *testing.T) {
	h, api := presetAPI(t)
	postLog(t, api, h.Bob, map[string]any{"date": day, "amount_oz": 8, "user_id": h.Alice.ID})
	var owner uint
	if err := h.Owner.Raw("SELECT user_id FROM water_logs").Scan(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if owner != h.Bob.ID {
		t.Fatalf("log owned by %d, want Bob (%d)", owner, h.Bob.ID)
	}
}

func TestWaterLogs_rlsBackstop(t *testing.T) {
	h, api := presetAPI(t)
	postLog(t, api, h.Alice, map[string]any{"date": day, "amount_oz": 16})

	err := h.Run(h.Bob, func(ctx context.Context) error {
		var rows []WaterLog
		if err := database.From(ctx).Unscoped().Find(&rows).Error; err != nil { // no Owned scope
			return err
		}
		if len(rows) != 0 {
			t.Errorf("unscoped query as Bob returned %d of Alice's rows", len(rows))
		}
		if n := database.From(ctx).Exec("UPDATE water_logs SET amount_oz = 1").RowsAffected; n != 0 {
			t.Errorf("unscoped update as Bob touched %d rows", n)
		}
		if n := database.From(ctx).Exec("DELETE FROM water_logs").RowsAffected; n != 0 {
			t.Errorf("unscoped delete as Bob touched %d rows", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	err = h.Run(h.Bob, func(ctx context.Context) error {
		return database.From(ctx).Exec("INSERT INTO water_logs (date, amount_oz, user_id) VALUES (now(), 1, ?)", h.Alice.ID).Error
	})
	if err == nil {
		t.Fatal("insert of a row for Alice as Bob succeeded; want a row-level security error")
	}
	var n int64
	if err := h.App.Model(&WaterLog{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("app role without app.user_id sees %d logs", n)
	}
}
