package controller_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/workout"
	"be-simpletracker/internal/core/workout/models"
	workoutrepo "be-simpletracker/internal/core/workout/repository"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Two-user tests for the workout module (docs/multi-user/DECISIONS.md "C3 workout"): exercises are
// a shared catalog, everything else is each user's own. Postgres only (TEST_DATABASE_URL).

func workoutAPI(t *testing.T) (*twousers.Harness, *twousers.API) {
	h := twousers.New(t)
	api := h.Engine(func(r *gin.Engine, auth gin.HandlerFunc) { workout.RegisterRoutes(r, auth) })
	return h, api
}

type idJSON struct {
	ID uint `json:"ID"`
}

type exerciseJSON struct {
	ID              uint                    `json:"ID"`
	Name            string                  `json:"name"`
	CreatedBy       uint                    `json:"created_by"`
	CanEdit         bool                    `json:"can_edit"`
	MachineSettings []models.MachineSetting `json:"machine_settings"`
}

func createExercise(t *testing.T, api *twousers.API, u twousers.User, name string, extra map[string]any) exerciseJSON {
	t.Helper()
	body := map[string]any{"name": name, "rep_rollover": 10}
	for k, v := range extra {
		body[k] = v
	}
	var out struct{ Exercise exerciseJSON }
	api.Do(u, http.MethodPost, "/workout/exercises", body).Status(t, http.StatusOK).JSON(t, &out)
	return out.Exercise
}

func listExercises(t *testing.T, api *twousers.API, u twousers.User) map[uint]exerciseJSON {
	t.Helper()
	var out struct{ Exercises []exerciseJSON }
	api.Do(u, http.MethodGet, "/workout/exercises/all", nil).Status(t, http.StatusOK).JSON(t, &out)
	m := map[uint]exerciseJSON{}
	for _, e := range out.Exercises {
		m[e.ID] = e
	}
	return m
}

type planJSON struct {
	ID        uint           `json:"ID"`
	Name      string         `json:"name"`
	Exercises []exerciseJSON `json:"exercises"`
}

type programJSON struct {
	ID       uint       `json:"ID"`
	Name     string     `json:"name"`
	IsActive bool       `json:"is_active"`
	Plans    []planJSON `json:"plans"`
}

func programs(t *testing.T, api *twousers.API, u twousers.User) []programJSON {
	t.Helper()
	var out struct{ Programs []programJSON }
	api.Do(u, http.MethodGet, "/workout/programs", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out.Programs
}

func newPlan(t *testing.T, api *twousers.API, u twousers.User, name string) planJSON {
	t.Helper()
	progs := programs(t, api, u)
	if len(progs) == 0 { // a new account has none: the first one made is the active one
		api.Do(u, http.MethodPost, "/workout/programs", map[string]any{"name": "Program"}).Status(t, http.StatusCreated)
		progs = programs(t, api, u)
	}
	var out struct{ Plan planJSON }
	api.Do(u, http.MethodPost, fmt.Sprintf("/workout/programs/%d/plans", progs[0].ID), map[string]any{"name": name}).
		Status(t, http.StatusCreated).JSON(t, &out)
	return out.Plan
}

func today() string { return clock.ZerodTime(0).Format("2006-01-02") }

type setJSON struct {
	ID       uint    `json:"ID"`
	ClientID string  `json:"client_id"`
	Reps     uint    `json:"reps"`
	Weight   float32 `json:"weight"`
}

type loggedJSON struct {
	ID           uint      `json:"ID"`
	WorkoutLogID uint      `json:"workout_log_id"`
	ExerciseID   uint      `json:"exercise_id"`
	Notes        string    `json:"notes"`
	Sets         []setJSON `json:"sets"`
}

func logSet(t *testing.T, api *twousers.API, u twousers.User, exerciseID uint, clientID string, want int) loggedJSON {
	t.Helper()
	var out struct {
		Exercise loggedJSON `json:"exercise"`
		Set      setJSON    `json:"set"`
		Created  bool       `json:"created"`
	}
	api.Do(u, http.MethodPost, "/workout/sets", map[string]any{
		"client_id": clientID, "date": today(), "exercise_id": exerciseID, "reps": 5, "weight": 100,
	}).Status(t, want).JSON(t, &out)
	return out.Exercise
}

type dayJSON struct {
	ID       uint         `json:"ID"`
	Exercise []loggedJSON `json:"exercises"`
	Cardio   *struct {
		Minutes int    `json:"minutes"`
		Type    string `json:"type"`
	} `json:"cardio"`
}

func getToday(t *testing.T, api *twousers.API, u twousers.User) dayJSON {
	t.Helper()
	var out dayJSON
	api.Do(u, http.MethodGet, "/workout/logs/today", nil).Status(t, http.StatusOK).JSON(t, &out)
	return out
}

const (
	aliceSet = "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a01"
	bobSet   = "0193a1b2-c3d4-7e5f-8a9b-0c1d2e3f4a02"
)

func count(t *testing.T, h *twousers.Harness, q string, args ...any) int64 {
	t.Helper()
	var n int64
	if err := h.Owner.Raw(q, args...).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestWorkout_requiresLogin(t *testing.T) {
	_, api := workoutAPI(t)
	for _, p := range []string{"/workout/programs", "/workout/plans/all", "/workout/exercises/all", "/workout/logs/today", "/workout/logs/month"} {
		api.Do(twousers.Anonymous, http.MethodGet, p, nil).Status(t, http.StatusUnauthorized)
	}
}

// Programs and plans are each user's own; a new user starts with none.
func TestWorkout_plansAreOwned(t *testing.T) {
	h, api := workoutAPI(t)
	squat := createExercise(t, api, h.Alice, "Squat", nil)
	plan := newPlan(t, api, h.Alice, "Alice leg day")
	api.Do(h.Alice, http.MethodPost, fmt.Sprintf("/workout/plans/%d/exercises/add", plan.ID), map[string]any{"exercise_id": squat.ID}).Status(t, http.StatusOK)

	if bobsPrograms := programs(t, api, h.Bob); len(bobsPrograms) != 0 {
		t.Fatalf("Bob's programs: %+v", bobsPrograms)
	}
	alicesProgram := programs(t, api, h.Alice)[0]
	var all struct{ Plans []planJSON }
	api.Do(h.Bob, http.MethodGet, "/workout/plans/all", nil).Status(t, http.StatusOK).JSON(t, &all)
	if len(all.Plans) != 0 {
		t.Fatalf("Bob sees plans: %+v", all.Plans)
	}

	// Every plan endpoint: Alice's plan and a missing one answer the same 404 for Bob.
	for _, id := range []uint{plan.ID, plan.ID + 1000} {
		p := fmt.Sprintf("/workout/plans/%d", id)
		api.Do(h.Bob, http.MethodPatch, p, map[string]any{"name": "mine now"}).NotFound(t)
		api.Do(h.Bob, http.MethodDelete, p, nil).NotFound(t)
		api.Do(h.Bob, http.MethodPost, p+"/restore", nil).NotFound(t)
		api.Do(h.Bob, http.MethodPost, p+"/exercises/add", map[string]any{"exercise_id": squat.ID}).NotFound(t)
		api.Do(h.Bob, http.MethodDelete, p+"/exercises/remove", map[string]any{"exercise_id": squat.ID}).NotFound(t)
		api.Do(h.Bob, http.MethodPut, p+"/exercises/reorder", map[string]any{"exercise_ids": []uint{squat.ID}}).Status(t, http.StatusBadRequest)
		api.Do(h.Bob, http.MethodPost, p+"/assign-day", map[string]any{"day_of_week": 2}).NotFound(t)
		api.Do(h.Bob, http.MethodDelete, p+"/assign-day", nil).NotFound(t)
		api.Do(h.Bob, http.MethodPut, p+"/planned-cardio", map[string]any{"type": "Run", "minutes": 20}).NotFound(t)
		api.Do(h.Bob, http.MethodPut, p+"/planned-mobility", map[string]any{"pre_mobility_items": []string{"x"}}).NotFound(t)
	}
	// Programs: same.
	for _, id := range []uint{alicesProgram.ID, alicesProgram.ID + 1000} {
		p := fmt.Sprintf("/workout/programs/%d", id)
		api.Do(h.Bob, http.MethodPatch, p, map[string]any{"name": "mine"}).NotFound(t)
		api.Do(h.Bob, http.MethodPost, p+"/activate", nil).NotFound(t)
		api.Do(h.Bob, http.MethodPost, p+"/plans", map[string]any{"name": "sneaky"}).NotFound(t)
	}
	// Bob's reorder of someone else's plan must not have reached the database.
	if n := count(t, h, `SELECT count(*) FROM workout_plans WHERE user_id = ?`, h.Bob.ID); n != 0 {
		t.Fatalf("Bob owns %d plans", n)
	}

	// Alice's plan is untouched.
	progs := programs(t, api, h.Alice)
	if len(progs) != 1 || len(progs[0].Plans) != 1 || progs[0].Plans[0].Name != "Alice leg day" ||
		len(progs[0].Plans[0].Exercises) != 1 || !progs[0].IsActive {
		t.Fatalf("Alice's program changed: %+v", progs)
	}
}

// A plan can't be given another user's exercise row, and Bob's plan is his: the composite FKs and
// RLS refuse what the handlers would have to be trusted to refuse.
func TestWorkout_compositeForeignKeys(t *testing.T) {
	h, api := workoutAPI(t)
	squat := createExercise(t, api, h.Alice, "Squat", nil)
	plan := newPlan(t, api, h.Alice, "Alice")
	api.Do(h.Alice, http.MethodPost, fmt.Sprintf("/workout/plans/%d/exercises/add", plan.ID), map[string]any{"exercise_id": squat.ID}).Status(t, http.StatusOK)
	logged := logSet(t, api, h.Alice, squat.ID, aliceSet, http.StatusOK)

	// The owner role bypasses RLS, so only the foreign keys stand in the way.
	for name, sql := range map[string]string{
		"plan day under Alice's plan, owned by Bob":      fmt.Sprintf(`INSERT INTO workout_plan_days (workout_plan_id, day_of_week, user_id) VALUES (%d, 3, %d)`, plan.ID, h.Bob.ID),
		"plan exercise under Alice's plan, owned by Bob": fmt.Sprintf(`INSERT INTO workout_plan_exercises (workout_plan_id, exercise_id, user_id) VALUES (%d, %d, %d)`, plan.ID, squat.ID+0, h.Bob.ID),
		"workout log pointing at Alice's plan":           fmt.Sprintf(`INSERT INTO workout_logs (date, workout_plan_id, user_id) VALUES ('2026-02-01', %d, %d)`, plan.ID, h.Bob.ID),
		"logged exercise under Alice's log":              fmt.Sprintf(`INSERT INTO logged_exercises (workout_log_id, exercise_id, user_id) VALUES (%d, %d, %d)`, logged.WorkoutLogID, squat.ID, h.Bob.ID),
		"set under Alice's logged exercise":              fmt.Sprintf(`INSERT INTO logged_sets (logged_exercise_id, reps, user_id) VALUES (%d, 5, %d)`, logged.ID, h.Bob.ID),
		"cardio under Alice's log":                       fmt.Sprintf(`INSERT INTO cardios (workout_log_id, minutes, user_id) VALUES (%d, 5, %d)`, logged.WorkoutLogID, h.Bob.ID),
		"plan in Alice's program, owned by Bob":          fmt.Sprintf(`INSERT INTO workout_plans (name, workout_program_id, user_id) VALUES ('x', %d, %d)`, programs(t, api, h.Alice)[0].ID, h.Bob.ID),
	} {
		if err := h.Owner.Exec(sql).Error; err == nil {
			t.Errorf("%s: insert succeeded, want a foreign key error", name)
		}
	}
}

// Both users start a workout on the same date and each gets their own log; neither blocks the other.
func TestWorkout_sameDateEachGetsTheirOwnLog(t *testing.T) {
	h, api := workoutAPI(t)
	squat := createExercise(t, api, h.Alice, "Squat", nil)

	a := getToday(t, api, h.Alice)
	b := getToday(t, api, h.Bob)
	if a.ID == 0 || b.ID == 0 || a.ID == b.ID {
		t.Fatalf("logs: alice %d, bob %d", a.ID, b.ID)
	}
	if n := count(t, h, `SELECT count(*) FROM workout_logs WHERE deleted_at IS NULL`); n != 2 {
		t.Fatalf("%d logs, want 2", n)
	}
	// Alice's first set doesn't hide Bob's log or make his start skip.
	logSet(t, api, h.Alice, squat.ID, aliceSet, http.StatusOK)
	if got := getToday(t, api, h.Bob); got.ID != b.ID || len(got.Exercise) != 0 {
		t.Fatalf("Bob's day after Alice logged: %+v", got)
	}
	// The same client_id is free for Bob: idempotency keys are per user.
	bob := logSet(t, api, h.Bob, squat.ID, aliceSet, http.StatusOK)
	if bob.WorkoutLogID != b.ID || len(bob.Sets) != 1 {
		t.Fatalf("Bob's set with Alice's client_id: %+v", bob)
	}
	if n := count(t, h, `SELECT count(*) FROM logged_sets WHERE client_id = ?`, aliceSet); n != 2 {
		t.Fatalf("%d sets with that client_id, want one per user", n)
	}
}

// BACKLOG #2: Alice's open transaction holding her date lock must not make Bob wait.
func TestWorkout_bobDoesNotWaitOnAlicesLock(t *testing.T) {
	h, api := workoutAPI(t)
	ctx, end := h.Ctx(h.Alice)
	log := models.WorkoutLog{Date: clock.ZerodTime(0)}
	if err := workoutrepo.CreateMinimalIfMissing(ctx, &log); err != nil || log.ID == 0 {
		t.Fatalf("Alice's create: id %d, %v", log.ID, err)
	}
	// Alice's transaction is still open: her lock and her uncommitted log are held.
	done := make(chan dayJSON, 1)
	go func() { done <- getToday(t, api, h.Bob) }()
	select {
	case got := <-done:
		if got.ID == 0 || got.ID == log.ID {
			t.Fatalf("Bob's day: %+v (Alice's log %d)", got, log.ID)
		}
	case <-time.After(5 * time.Second):
		_ = end(false)
		t.Fatal("Bob waited on Alice's lock")
	}
	if err := end(true); err != nil {
		t.Fatal(err)
	}
}

func TestWorkout_logsAreOwned(t *testing.T) {
	h, api := workoutAPI(t)
	squat := createExercise(t, api, h.Alice, "Squat", nil)
	alices := logSet(t, api, h.Alice, squat.ID, aliceSet, http.StatusOK)
	set := alices.Sets[0]
	api.Do(h.Alice, http.MethodPost, "/workout/logs/cardio", map[string]any{"minutes": 20, "type": "Run"}).Status(t, http.StatusOK)

	// Bob's views are his own and empty.
	b := getToday(t, api, h.Bob)
	if len(b.Exercise) != 0 || b.Cardio != nil {
		t.Fatalf("Bob's day shows Alice's data: %+v", b)
	}
	var month struct{ Days []dayJSON }
	api.Do(h.Bob, http.MethodGet, "/workout/logs/month", nil).Status(t, http.StatusOK).JSON(t, &month)
	for _, d := range month.Days {
		if len(d.Exercise) != 0 {
			t.Fatalf("Bob's month has Alice's exercises: %+v", d)
		}
	}
	var act struct {
		ActiveDates []string `json:"active_dates"`
	}
	api.Do(h.Bob, http.MethodGet, "/workout/logs/activity", nil).Status(t, http.StatusOK).JSON(t, &act)
	if len(act.ActiveDates) != 0 {
		t.Fatalf("Bob's activity: %v", act.ActiveDates)
	}
	var prog struct{ Progression []any }
	api.Do(h.Bob, http.MethodGet, fmt.Sprintf("/workout/exercises/progression/%d", squat.ID), nil).Status(t, http.StatusOK).JSON(t, &prog)
	if len(prog.Progression) != 0 {
		t.Fatalf("Bob's progression: %v", prog.Progression)
	}
	api.Do(h.Alice, http.MethodGet, fmt.Sprintf("/workout/exercises/progression/%d", squat.ID), nil).Status(t, http.StatusOK).JSON(t, &prog)
	if len(prog.Progression) != 1 {
		t.Fatalf("Alice's progression: %v", prog.Progression)
	}

	// Sets: Alice's and a missing one answer the same for Bob (by id and by client_id).
	for _, ref := range []string{fmt.Sprint(set.ID), aliceSet, fmt.Sprint(set.ID + 1000), bobSet} {
		api.Do(h.Bob, http.MethodPatch, "/workout/sets/"+ref, map[string]any{"reps": 99}).NotFound(t)
	}
	api.Do(h.Bob, http.MethodDelete, fmt.Sprintf("/workout/exercises/sets/%d", set.ID), nil).NotFound(t)
	var del struct{ Deleted bool }
	api.Do(h.Bob, http.MethodDelete, fmt.Sprintf("/workout/sets/%d", set.ID), nil).Status(t, http.StatusOK).JSON(t, &del)
	if del.Deleted {
		t.Fatal("Bob deleted Alice's set")
	}

	// Legacy full-replace endpoint: can't write into Alice's workout log or logged exercise.
	for _, typ := range []string{"previous", "logged"} {
		api.Do(h.Bob, http.MethodPost, "/workout/exercises/log", map[string]any{
			"type": typ,
			"exercise": map[string]any{
				"ID": alices.ID, "workout_log_id": alices.WorkoutLogID, "exercise_id": squat.ID, "notes": "mine",
				"sets": []map[string]any{{"reps": 1, "weight": 1}},
			},
		}).NotFound(t)
	}
	// Add / remove exercise act on Bob's own day only.
	var added struct{ Exercise loggedJSON }
	api.Do(h.Bob, http.MethodPost, "/workout/exercises/add", map[string]any{"exercise_id": squat.ID}).Status(t, http.StatusOK).JSON(t, &added)
	if added.Exercise.WorkoutLogID != b.ID {
		t.Fatalf("exercise added to log %d, Bob's is %d", added.Exercise.WorkoutLogID, b.ID)
	}
	api.Do(h.Bob, http.MethodDelete, "/workout/exercises/remove", map[string]any{"exercise_id": squat.ID}).Status(t, http.StatusOK)

	// Alice's data is all still there.
	a := getToday(t, api, h.Alice)
	if len(a.Exercise) != 1 || len(a.Exercise[0].Sets) != 1 || a.Exercise[0].Sets[0].Reps != 5 || a.Cardio == nil || a.Cardio.Minutes != 20 {
		t.Fatalf("Alice's day changed: %+v", a)
	}
}

func TestWorkout_cardioIsPerUser(t *testing.T) {
	h, api := workoutAPI(t)
	api.Do(h.Alice, http.MethodPost, "/workout/logs/cardio", map[string]any{"minutes": 20, "type": "Run"}).Status(t, http.StatusOK)
	api.Do(h.Bob, http.MethodPost, "/workout/logs/cardio", map[string]any{"minutes": 45, "type": "Bike"}).Status(t, http.StatusOK)
	a, b := getToday(t, api, h.Alice), getToday(t, api, h.Bob)
	if a.Cardio == nil || a.Cardio.Minutes != 20 || a.Cardio.Type != "Run" || b.Cardio == nil || b.Cardio.Minutes != 45 || b.Cardio.Type != "Bike" {
		t.Fatalf("cardio: alice %+v, bob %+v", a.Cardio, b.Cardio)
	}
}

// Forged user_id / created_by in any body is ignored: the caller is the owner.
func TestWorkout_forgedOwnerFieldsAreIgnored(t *testing.T) {
	h, api := workoutAPI(t)
	forged := map[string]any{"user_id": h.Alice.ID, "UserID": h.Alice.ID, "created_by": h.Alice.ID, "CreatedBy": h.Alice.ID}
	with := func(m map[string]any) map[string]any {
		for k, v := range forged {
			m[k] = v
		}
		return m
	}

	bobsEx := createExercise(t, api, h.Bob, "Bob's curl", forged)
	if bobsEx.CreatedBy != h.Bob.ID {
		t.Fatalf("created_by = %d, want Bob %d", bobsEx.CreatedBy, h.Bob.ID)
	}
	var prog struct{ Program programJSON }
	api.Do(h.Bob, http.MethodPost, "/workout/programs", with(map[string]any{"name": "Bob's"})).Status(t, http.StatusCreated).JSON(t, &prog)
	api.Do(h.Bob, http.MethodPost, fmt.Sprintf("/workout/programs/%d/plans", prog.Program.ID), with(map[string]any{"name": "Bob's plan"})).Status(t, http.StatusCreated)
	logged := logSet(t, api, h.Bob, bobsEx.ID, bobSet, http.StatusOK)
	api.Do(h.Bob, http.MethodPost, "/workout/exercises/log", map[string]any{
		"type": "logged",
		"exercise": with(map[string]any{
			"ID": logged.ID, "workout_log_id": logged.WorkoutLogID, "exercise_id": bobsEx.ID,
			"sets": []map[string]any{with(map[string]any{"reps": 7, "weight": 10})},
		}),
	}).Status(t, http.StatusOK)
	api.Do(h.Bob, http.MethodPost, "/workout/logs/cardio", with(map[string]any{"minutes": 5, "type": "Row"})).Status(t, http.StatusOK)

	for _, tbl := range []string{"workout_programs", "workout_plans", "workout_logs", "logged_exercises", "logged_sets", "cardios"} {
		if n := count(t, h, fmt.Sprintf(`SELECT count(*) FROM %s WHERE user_id = ?`, tbl), h.Alice.ID); n != 0 {
			t.Errorf("%s: %d rows belong to Alice after Bob's requests", tbl, n)
		}
		if n := count(t, h, fmt.Sprintf(`SELECT count(*) FROM %s WHERE user_id = ?`, tbl), h.Bob.ID); n == 0 {
			t.Errorf("%s: no rows belong to Bob", tbl)
		}
	}
	if n := count(t, h, `SELECT count(*) FROM exercises WHERE created_by = ?`, h.Alice.ID); n != 0 {
		t.Fatalf("%d exercises credited to Alice", n)
	}
}

// Exercises are a shared catalog: anyone adds, the creator or the admin edits, only the admin deletes.
func TestWorkout_exerciseCatalogPermissions(t *testing.T) {
	h, api := workoutAPI(t)
	ex := createExercise(t, api, h.Alice, "Alice's lift", nil)
	path := fmt.Sprintf("/workout/exercises/%d", ex.ID)
	edit := map[string]any{"name": "Renamed", "rep_rollover": 8, "cues": "new cue", "load_type": "weight_stack"}

	// Everyone reads it, with can_edit for the creator and the admin only.
	if got := listExercises(t, api, h.Bob)[ex.ID]; got.Name != "Alice's lift" || got.CanEdit || got.CreatedBy != h.Alice.ID {
		t.Fatalf("Bob's view: %+v", got)
	}
	if got := listExercises(t, api, h.Alice)[ex.ID]; !got.CanEdit {
		t.Fatalf("Alice's view: %+v", got)
	}
	if got := listExercises(t, api, h.Admin)[ex.ID]; !got.CanEdit {
		t.Fatalf("admin's view: %+v", got)
	}

	// Bob can neither edit nor delete: 403 (the row is visible to him, so 403 leaks nothing).
	api.Do(h.Bob, http.MethodPut, path, edit).Status(t, http.StatusForbidden)
	api.Do(h.Bob, http.MethodPut, path+"/cues", map[string]any{"cues": "x"}).Status(t, http.StatusForbidden)
	api.Do(h.Bob, http.MethodDelete, path, nil).Status(t, http.StatusForbidden)
	api.Do(h.Alice, http.MethodDelete, path, nil).Status(t, http.StatusForbidden)
	// A missing exercise is a plain 404 for everyone.
	api.Do(h.Bob, http.MethodPut, fmt.Sprintf("/workout/exercises/%d", ex.ID+1000), edit).NotFound(t)
	api.Do(h.Admin, http.MethodDelete, fmt.Sprintf("/workout/exercises/%d", ex.ID+1000), nil).NotFound(t)
	if got := listExercises(t, api, h.Alice)[ex.ID]; got.Name != "Alice's lift" {
		t.Fatalf("Bob's rejected edit changed the exercise: %+v", got)
	}

	// The creator edits; so does the admin.
	api.Do(h.Alice, http.MethodPut, path, edit).Status(t, http.StatusOK)
	edit["name"] = "Admin's rename"
	api.Do(h.Admin, http.MethodPut, path, edit).Status(t, http.StatusOK)
	api.Do(h.Admin, http.MethodPut, path+"/cues", map[string]any{"cues": "admin cue"}).Status(t, http.StatusOK)
	if got := listExercises(t, api, h.Bob)[ex.ID]; got.Name != "Admin's rename" || got.CreatedBy != h.Alice.ID {
		t.Fatalf("after edits: %+v", got)
	}

	// Only the admin deletes, and not while her own plan still uses it.
	plan := newPlan(t, api, h.Admin, "Admin plan")
	api.Do(h.Admin, http.MethodPost, fmt.Sprintf("/workout/plans/%d/exercises/add", plan.ID), map[string]any{"exercise_id": ex.ID}).Status(t, http.StatusOK)
	api.Do(h.Admin, http.MethodDelete, path, nil).Status(t, http.StatusConflict)
	api.Do(h.Admin, http.MethodDelete, fmt.Sprintf("/workout/plans/%d/exercises/remove", plan.ID), map[string]any{"exercise_id": ex.ID}).Status(t, http.StatusOK)
	api.Do(h.Admin, http.MethodDelete, path, nil).Status(t, http.StatusNoContent)
	if _, ok := listExercises(t, api, h.Bob)[ex.ID]; ok {
		t.Fatal("deleted exercise still listed")
	}
}

// Each user's machine settings are their own; the shared exercise carries none.
func TestWorkout_machineSettingsArePerUser(t *testing.T) {
	h, api := workoutAPI(t)
	ex := createExercise(t, api, h.Alice, "Chest press", nil)
	path := fmt.Sprintf("/workout/exercises/%d/settings", ex.ID)
	save := func(u twousers.User, settings []map[string]string, want int) {
		api.Do(u, http.MethodPut, path, map[string]any{"machine_settings": settings}).Status(t, want)
	}
	settingsOf := func(u twousers.User) []models.MachineSetting { return listExercises(t, api, u)[ex.ID].MachineSettings }

	save(h.Alice, []map[string]string{{"label": "Seat", "value": "4"}, {"label": "Pad", "value": "2"}}, http.StatusOK)
	if got := settingsOf(h.Alice); len(got) != 2 || got[0].Label != "Seat" || got[0].Value != "4" {
		t.Fatalf("Alice's settings: %+v", got)
	}
	if got := settingsOf(h.Bob); len(got) != 0 {
		t.Fatalf("Bob sees Alice's settings: %+v", got)
	}
	// Bob, who can't edit the exercise, still saves his own; Alice's stay put.
	save(h.Bob, []map[string]string{{"label": "Seat", "value": "7"}}, http.StatusOK)
	if got := settingsOf(h.Bob); len(got) != 1 || got[0].Value != "7" {
		t.Fatalf("Bob's settings: %+v", got)
	}
	if got := settingsOf(h.Alice); len(got) != 2 || got[0].Value != "4" {
		t.Fatalf("Alice's settings after Bob saved: %+v", got)
	}
	// They also show where the app puts them: a routine's exercises.
	plan := newPlan(t, api, h.Bob, "Bob's routine")
	var withEx struct{ Plan planJSON }
	api.Do(h.Bob, http.MethodPost, fmt.Sprintf("/workout/plans/%d/exercises/add", plan.ID), map[string]any{"exercise_id": ex.ID}).Status(t, http.StatusOK).JSON(t, &withEx)
	if got := withEx.Plan.Exercises; len(got) != 1 || len(got[0].MachineSettings) != 1 || got[0].MachineSettings[0].Value != "7" {
		t.Fatalf("plan exercise settings: %+v", got)
	}
	// Replace, clear, validate.
	save(h.Bob, []map[string]string{{"label": "Seat", "value": "8"}, {"label": " ", "value": " "}}, http.StatusOK)
	if got := settingsOf(h.Bob); len(got) != 1 || got[0].Value != "8" {
		t.Fatalf("Bob's settings after replace: %+v", got)
	}
	save(h.Bob, nil, http.StatusOK)
	if got := settingsOf(h.Bob); len(got) != 0 {
		t.Fatalf("Bob's settings after clear: %+v", got)
	}
	save(h.Bob, []map[string]string{{"label": "", "value": "3"}}, http.StatusBadRequest)
	api.Do(h.Bob, http.MethodPut, fmt.Sprintf("/workout/exercises/%d/settings", ex.ID+1000), map[string]any{"machine_settings": []any{}}).NotFound(t)
	if got := settingsOf(h.Alice); len(got) != 2 {
		t.Fatalf("Alice's settings after Bob's changes: %+v", got)
	}
	if n := count(t, h, `SELECT count(*) FROM exercise_settings`); n != 1 {
		t.Fatalf("%d settings rows, want only Alice's", n)
	}
}

// RLS under the handlers: queries that forget every scope still see only the caller's rows.
func TestWorkout_rlsBackstop(t *testing.T) {
	h, api := workoutAPI(t)
	squat := createExercise(t, api, h.Alice, "Squat", nil)
	plan := newPlan(t, api, h.Alice, "Alice")
	api.Do(h.Alice, http.MethodPost, fmt.Sprintf("/workout/plans/%d/exercises/add", plan.ID), map[string]any{"exercise_id": squat.ID}).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodPost, fmt.Sprintf("/workout/plans/%d/assign-day", plan.ID), map[string]any{"day_of_week": 2}).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodPut, fmt.Sprintf("/workout/exercises/%d/settings", squat.ID), map[string]any{"machine_settings": []map[string]string{{"label": "Seat", "value": "4"}}}).Status(t, http.StatusOK)
	logged := logSet(t, api, h.Alice, squat.ID, aliceSet, http.StatusOK)
	api.Do(h.Alice, http.MethodPost, "/workout/logs/cardio", map[string]any{"minutes": 20, "type": "Run"}).Status(t, http.StatusOK)

	owned := []string{"workout_programs", "workout_plans", "workout_plan_days", "workout_plan_exercises",
		"workout_logs", "logged_exercises", "logged_sets", "cardios", "exercise_settings"}
	err := h.Run(h.Bob, func(ctx context.Context) error {
		for _, tbl := range owned {
			var n int64
			if err := database.From(ctx).Table(tbl).Count(&n).Error; err != nil { // no Owned scope
				return err
			}
			if n != 0 {
				t.Errorf("unscoped count of %s as Bob: %d rows", tbl, n)
			}
			res := database.From(ctx).Exec(fmt.Sprintf("DELETE FROM %s", tbl))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected != 0 {
				t.Errorf("unscoped delete on %s as Bob touched %d rows", tbl, res.RowsAffected)
			}
		}
		// The catalog is readable but not editable by a non-creator.
		var ex []models.Exercise
		if err := database.From(ctx).Find(&ex).Error; err != nil || len(ex) != 1 {
			t.Errorf("Bob reads the catalog: %d rows, %v", len(ex), err)
		}
		if n := database.From(ctx).Exec("UPDATE exercises SET name = 'x'").RowsAffected; n != 0 {
			t.Errorf("Bob edited %d of Alice's exercises", n)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	// Writing rows for Alice as Bob is refused outright.
	for name, sql := range map[string]string{
		"workout_programs":  fmt.Sprintf(`INSERT INTO workout_programs (name, user_id) VALUES ('x', %d)`, h.Alice.ID),
		"workout_logs":      fmt.Sprintf(`INSERT INTO workout_logs (date, user_id) VALUES ('2026-03-01', %d)`, h.Alice.ID),
		"cardios":           fmt.Sprintf(`INSERT INTO cardios (workout_log_id, minutes, user_id) VALUES (%d, 1, %d)`, logged.WorkoutLogID, h.Alice.ID),
		"exercise_settings": fmt.Sprintf(`INSERT INTO exercise_settings (exercise_id, user_id) VALUES (%d, %d)`, squat.ID, h.Alice.ID),
		"exercises":         fmt.Sprintf(`INSERT INTO exercises (name, created_by) VALUES ('x', %d)`, h.Alice.ID),
	} {
		if err := h.Run(h.Bob, func(ctx context.Context) error { return database.From(ctx).Exec(sql).Error }); err == nil {
			t.Errorf("%s: inserting for Alice as Bob succeeded; want a row-level security error", name)
		}
	}
	// A soft delete is an UPDATE of deleted_at: catalog_guard refuses it for a non-admin.
	if err := h.Run(h.Alice, func(ctx context.Context) error {
		return database.From(ctx).Exec("UPDATE exercises SET deleted_at = now()").Error
	}); err == nil {
		t.Error("a non-admin soft-deleted an exercise")
	}
	if err := h.Run(h.Admin, func(ctx context.Context) error {
		return database.From(ctx).Exec("UPDATE exercises SET deleted_at = now()").Error
	}); err != nil {
		t.Errorf("admin soft delete: %v", err)
	}
}

// PATCH /workout/logs/switch-plan: Bob can't point his day at Alice's plan, and the 404 matches a missing plan.
func TestWorkout_switchPlanIsOwned(t *testing.T) {
	h, api := workoutAPI(t)
	alices := newPlan(t, api, h.Alice, "Alice's plan")
	bobs := newPlan(t, api, h.Bob, "Bob's plan")

	api.Do(h.Alice, http.MethodPatch, "/workout/logs/switch-plan", map[string]any{"plan_id": alices.ID}).Status(t, http.StatusOK)
	api.Do(h.Bob, http.MethodPatch, "/workout/logs/switch-plan", map[string]any{"plan_id": bobs.ID}).Status(t, http.StatusOK)

	missing := api.Do(h.Bob, http.MethodPatch, "/workout/logs/switch-plan", map[string]any{"plan_id": 999999}).Status(t, http.StatusNotFound)
	other := api.Do(h.Bob, http.MethodPatch, "/workout/logs/switch-plan", map[string]any{"plan_id": alices.ID}).Status(t, http.StatusNotFound)
	if string(missing.Body) != string(other.Body) {
		t.Fatalf("Alice's plan answers %s, a missing plan %s: they must match", other.Body, missing.Body)
	}
	if n := count(t, h, `SELECT count(*) FROM workout_logs WHERE user_id = ? AND workout_plan_id = ?`, h.Bob.ID, alices.ID); n != 0 {
		t.Fatalf("Bob's log points at Alice's plan")
	}
}
