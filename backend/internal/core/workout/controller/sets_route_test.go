package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"be-simpletracker/internal/core/workout"
	"be-simpletracker/internal/core/workout/models"
	"be-simpletracker/internal/core/workout/testutil"

	"github.com/gin-gonic/gin"
)

func setsRouter(t *testing.T) (*gin.Engine, models.Exercise) {
	t.Helper()
	db := testutil.SetupTestDB(t)
	ex := models.Exercise{Name: "Squat"}
	if err := db.Create(&ex).Error; err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	r := gin.New()
	workout.RegisterRoutes(r, func(c *gin.Context) { c.Next() })
	return r, ex
}

func do(t *testing.T, r *gin.Engine, method, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	var out map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func TestSetsRoutes_logReplayPatchDelete(t *testing.T) {
	r, ex := setsRouter(t)
	const cid = "0193A1B2-C3D4-7E5F-8A9B-0C1D2E3F4A5B" // upper case: normalized server-side
	body, _ := json.Marshal(map[string]any{
		"client_id": cid, "date": "2026-11-20", "exercise_id": ex.ID, "reps": 5, "weight": 100, "weight_setup": "",
	})

	code, first := do(t, r, http.MethodPost, "/workout/sets", string(body))
	if code != http.StatusOK || first["created"] != true || first["first_set_of_day"] != true {
		t.Fatalf("first POST: %d %v", code, first)
	}
	set := first["set"].(map[string]any)
	if set["client_id"] != "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b" {
		t.Fatalf("client_id not normalized: %v", set["client_id"])
	}
	exercise := first["exercise"].(map[string]any)
	if len(exercise["sets"].([]any)) != 1 {
		t.Fatalf("response should carry the full logged exercise: %v", exercise)
	}

	code, again := do(t, r, http.MethodPost, "/workout/sets", string(body))
	if code != http.StatusOK || again["created"] != false || again["set"].(map[string]any)["ID"] != set["ID"] {
		t.Fatalf("replay POST: %d %v", code, again)
	}

	code, patched := do(t, r, http.MethodPatch, "/workout/sets/"+cid, `{"reps": 6}`)
	if code != http.StatusOK || patched["set"].(map[string]any)["reps"] != float64(6) {
		t.Fatalf("PATCH: %d %v", code, patched)
	}

	code, del := do(t, r, http.MethodDelete, "/workout/sets/"+cid, "")
	if code != http.StatusOK || del["deleted"] != true || del["exercise_removed"] != true {
		t.Fatalf("DELETE: %d %v", code, del)
	}
	code, del = do(t, r, http.MethodDelete, "/workout/sets/"+cid, "")
	if code != http.StatusOK || del["deleted"] != false {
		t.Fatalf("repeat DELETE should be 200 deleted=false: %d %v", code, del)
	}
	code, _ = do(t, r, http.MethodPatch, "/workout/sets/"+cid, `{"reps": 7}`)
	if code != http.StatusNotFound {
		t.Fatalf("PATCH on a deleted set: want 404, got %d", code)
	}
}

func TestSetsRoutes_validation(t *testing.T) {
	r, _ := setsRouter(t)
	cases := []struct{ name, body string }{
		{"missing client_id", `{"exercise_id": 1, "reps": 5}`},
		{"bad client_id", `{"client_id": "abc", "exercise_id": 1}`},
		{"missing exercise", `{"client_id": "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b"}`},
		{"bad date", `{"client_id": "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b", "exercise_id": 1, "date": "20/11/2026"}`},
	}
	for _, tc := range cases {
		if code, _ := do(t, r, http.MethodPost, "/workout/sets", tc.body); code != http.StatusBadRequest {
			t.Errorf("%s: want 400, got %d", tc.name, code)
		}
	}
	if code, _ := do(t, r, http.MethodPost, "/workout/sets", `{"client_id": "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b", "exercise_id": 999}`); code != http.StatusNotFound {
		t.Errorf("unknown exercise: want 404, got %d", code)
	}
	if code, _ := do(t, r, http.MethodDelete, "/workout/sets/not-an-id", ""); code != http.StatusBadRequest {
		t.Errorf("bad set ref: want 400, got %d", code)
	}
}

// The legacy full-replace endpoint must not create a second parent on a double tap.
func TestLegacyLogExercise_doubleTapNoDuplicate(t *testing.T) {
	r, ex := setsRouter(t)
	if code, out := do(t, r, http.MethodGet, "/workout/logs/previous", ""); code != http.StatusOK {
		t.Fatalf("GET previous (creates today): %d %v", code, out)
	}
	// The legacy client sends workout_log_id from the day view; today's log is id 1 here.
	body, _ := json.Marshal(map[string]any{
		"type": "previous",
		"exercise": map[string]any{
			"workout_log_id": 1, "exercise_id": ex.ID,
			"sets": []map[string]any{{"reps": 5, "weight": 100}},
		},
	})
	for i := 0; i < 2; i++ {
		code, res := do(t, r, http.MethodPost, "/workout/exercises/log", string(body))
		if code != http.StatusOK {
			t.Fatalf("POST %d: %d %v", i, code, res)
		}
		sets := res["exercise"].(map[string]any)["sets"].([]any)
		if len(sets) != 1 {
			t.Fatalf("POST %d: want 1 set, got %d", i, len(sets))
		}
	}
}
