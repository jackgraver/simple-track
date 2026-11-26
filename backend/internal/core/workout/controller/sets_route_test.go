package controller_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
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
		{"missing exercise", `{"client_id": "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b", "date": "2026-11-20"}`},
		{"missing date", `{"client_id": "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b", "exercise_id": 1, "reps": 5}`},
		{"bad date", `{"client_id": "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b", "exercise_id": 1, "date": "20/11/2026"}`},
	}
	for _, tc := range cases {
		if code, _ := do(t, r, http.MethodPost, "/workout/sets", tc.body); code != http.StatusBadRequest {
			t.Errorf("%s: want 400, got %d", tc.name, code)
		}
	}
	if code, _ := do(t, r, http.MethodPost, "/workout/sets", `{"client_id": "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a5b", "exercise_id": 999, "date": "2026-11-20"}`); code != http.StatusNotFound {
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

func TestSetsRoutes_unilateralSides(t *testing.T) {
	r, ex := setsRouter(t)
	post := func(cid, side string) (int, map[string]any) {
		body := map[string]any{"client_id": cid, "date": "2026-11-20", "exercise_id": ex.ID, "reps": 8, "weight": 30}
		if side != "" {
			body["side"] = side
		}
		b, _ := json.Marshal(body)
		return do(t, r, http.MethodPost, "/workout/sets", string(b))
	}
	const left = "0193a1b2-c3d4-7e5f-8a9b-000000000001"
	const right = "0193a1b2-c3d4-7e5f-8a9b-000000000002"
	const both = "0193a1b2-c3d4-7e5f-8a9b-000000000003"
	if code, out := post(left, "left"); code != http.StatusOK || out["set"].(map[string]any)["side"] != "left" {
		t.Fatalf("left: %d %v", code, out)
	}
	if code, out := post(right, "right"); code != http.StatusOK || out["set"].(map[string]any)["side"] != "right" {
		t.Fatalf("right: %d %v", code, out)
	}
	if code, out := post(both, ""); code != http.StatusOK || out["set"].(map[string]any)["side"] != nil {
		t.Fatalf("no side should be bilateral (null): %d %v", code, out)
	}
	if code, _ := post("0193a1b2-c3d4-7e5f-8a9b-000000000004", "middle"); code != http.StatusBadRequest {
		t.Fatalf("bad side: want 400, got %d", code)
	}

	// pair_id is stored on sided sets, ignored on bilateral ones, and must be a uuid.
	postPair := func(cid, side, pair string) (int, map[string]any) {
		body := map[string]any{"client_id": cid, "date": "2026-11-20", "exercise_id": ex.ID, "reps": 8, "weight": 30, "pair_id": pair}
		if side != "" {
			body["side"] = side
		}
		b, _ := json.Marshal(body)
		return do(t, r, http.MethodPost, "/workout/sets", string(b))
	}
	const paired = "0193a1b2-c3d4-7e5f-8a9b-000000000005"
	if code, out := postPair(paired, "left", left); code != http.StatusOK || out["set"].(map[string]any)["pair_id"] != left {
		t.Fatalf("pair_id: %d %v", code, out)
	}
	if code, out := postPair("0193a1b2-c3d4-7e5f-8a9b-000000000006", "", left); code != http.StatusOK || out["set"].(map[string]any)["pair_id"] != nil {
		t.Fatalf("bilateral pair_id must be dropped: %d %v", code, out)
	}
	if code, _ := postPair("0193a1b2-c3d4-7e5f-8a9b-000000000007", "left", "nope"); code != http.StatusBadRequest {
		t.Fatalf("bad pair_id: want 400, got %d", code)
	}
	if code, out := do(t, r, http.MethodPatch, "/workout/sets/"+paired, `{"pair_id": "`+right+`"}`); code != http.StatusOK || out["set"].(map[string]any)["pair_id"] != right {
		t.Fatalf("PATCH pair_id: %d %v", code, out)
	}
	if code, out := do(t, r, http.MethodPatch, "/workout/sets/"+paired, `{"pair_id": ""}`); code != http.StatusOK || out["set"].(map[string]any)["pair_id"] != nil {
		t.Fatalf("PATCH pair_id clear: %d %v", code, out)
	}
	if code, _ := do(t, r, http.MethodPatch, "/workout/sets/"+paired, `{"pair_id": "x"}`); code != http.StatusBadRequest {
		t.Fatalf("PATCH bad pair_id: want 400, got %d", code)
	}

	if code, out := do(t, r, http.MethodPatch, "/workout/sets/"+left, `{"side": "right"}`); code != http.StatusOK || out["set"].(map[string]any)["side"] != "right" {
		t.Fatalf("PATCH side: %d %v", code, out)
	}
	if code, out := do(t, r, http.MethodPatch, "/workout/sets/"+left, `{"side": ""}`); code != http.StatusOK || out["set"].(map[string]any)["side"] != nil {
		t.Fatalf("PATCH side to bilateral: %d %v", code, out)
	}
	if code, out := do(t, r, http.MethodPatch, "/workout/sets/"+right, `{"reps": 9}`); code != http.StatusOK || out["set"].(map[string]any)["side"] != "right" {
		t.Fatalf("PATCH without side must keep it: %d %v", code, out)
	}
	if code, _ := do(t, r, http.MethodPatch, "/workout/sets/"+right, `{"side": "up"}`); code != http.StatusBadRequest {
		t.Fatalf("PATCH bad side: want 400, got %d", code)
	}
}

func TestUpdateExercise_unilateralToggle(t *testing.T) {
	r, ex := setsRouter(t)
	path := "/workout/exercises/" + strconv.FormatUint(uint64(ex.ID), 10)
	code, out := do(t, r, http.MethodPut, path, `{"name": "Squat", "rep_rollover": 10, "unilateral": true}`)
	if code != http.StatusOK || out["exercise"].(map[string]any)["unilateral"] != true {
		t.Fatalf("turn on: %d %v", code, out)
	}
	code, out = do(t, r, http.MethodPut, path, `{"name": "Squat", "rep_rollover": 10}`)
	if code != http.StatusOK || out["exercise"].(map[string]any)["unilateral"] != true {
		t.Fatalf("omitting unilateral must keep it: %d %v", code, out)
	}
	code, out = do(t, r, http.MethodPut, path, `{"name": "Squat", "rep_rollover": 10, "unilateral": false}`)
	if code != http.StatusOK || out["exercise"].(map[string]any)["unilateral"] != false {
		t.Fatalf("turn off: %d %v", code, out)
	}
}
