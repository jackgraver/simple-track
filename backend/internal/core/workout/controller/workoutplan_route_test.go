package controller

// import (
// 	"bytes"
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"strconv"
// 	"sync/atomic"
// 	"testing"

// 	"be-simpletracker/internal/core/workout/models"

// 	"github.com/gin-gonic/gin"
// 	"github.com/glebarez/sqlite"
// 	"gorm.io/gorm"
// 	"gorm.io/gorm/logger"
// )

// var workoutPlanTestDBCounter atomic.Uint64

// func workoutPlanTestDB(t *testing.T) *gorm.DB {
// 	t.Helper()
// 	id := workoutPlanTestDBCounter.Add(1)
// 	dsn := "file:workoutplan_route_test_" + strconv.FormatUint(id, 10) + "?mode=memory&cache=shared"
// 	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
// 		Logger: logger.Default.LogMode(logger.Silent),
// 	})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if err := db.AutoMigrate(
// 		&models.WorkoutPlan{},
// 		&models.Exercise{},
// 		&models.WorkoutPlanExercise{},
// 	); err != nil {
// 		t.Fatal(err)
// 	}
// 	return db
// }

// func workoutPlanTestRouter(t *testing.T, db *gorm.DB) *gin.Engine {
// 	t.Helper()
// 	gin.SetMode(gin.TestMode)
// 	r := gin.New()
// 	RegisterWorkoutPlanRoutes(r.Group("/workout"), db)
// 	return r
// }

// func TestWorkoutPlanHandler_assignPlanToDay_dayZeroAccepted(t *testing.T) {
// 	db := workoutPlanTestDB(t)
// 	plan := models.WorkoutPlan{Name: "Legs"}
// 	if err := db.Create(&plan).Error; err != nil {
// 		t.Fatal(err)
// 	}
// 	router := workoutPlanTestRouter(t, db)

// 	req := httptest.NewRequest(http.MethodPost, "/workout/plans/1/assign-day", bytes.NewBufferString(`{"day_of_week":0}`))
// 	req.Header.Set("Content-Type", "application/json")
// 	rec := httptest.NewRecorder()
// 	router.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusOK {
// 		t.Fatalf("status %d, body: %s", rec.Code, rec.Body.String())
// 	}
// 	var body struct {
// 		Plan struct {
// 			DayOfWeek *int `json:"day_of_week"`
// 		} `json:"plan"`
// 	}
// 	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
// 		t.Fatal(err)
// 	}
// 	if body.Plan.DayOfWeek == nil {
// 		t.Fatal("expected day_of_week in response")
// 	}
// 	if *body.Plan.DayOfWeek != 0 {
// 		t.Fatalf("day_of_week got %d want 0", *body.Plan.DayOfWeek)
// 	}
// }

// func TestWorkoutPlanHandler_assignPlanToDay_invalidPlanID(t *testing.T) {
// 	db := workoutPlanTestDB(t)
// 	router := workoutPlanTestRouter(t, db)

// 	req := httptest.NewRequest(http.MethodPost, "/workout/plans/not-an-id/assign-day", bytes.NewBufferString(`{"day_of_week":1}`))
// 	req.Header.Set("Content-Type", "application/json")
// 	rec := httptest.NewRecorder()
// 	router.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusBadRequest {
// 		t.Fatalf("status %d, body: %s", rec.Code, rec.Body.String())
// 	}
// }

// func TestWorkoutPlanHandler_assignPlanToDay_planNotFound(t *testing.T) {
// 	db := workoutPlanTestDB(t)
// 	router := workoutPlanTestRouter(t, db)

// 	req := httptest.NewRequest(http.MethodPost, "/workout/plans/99/assign-day", bytes.NewBufferString(`{"day_of_week":2}`))
// 	req.Header.Set("Content-Type", "application/json")
// 	rec := httptest.NewRecorder()
// 	router.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusInternalServerError {
// 		t.Fatalf("status %d, body: %s", rec.Code, rec.Body.String())
// 	}
// }

// func TestWorkoutPlanHandler_assignPlanToDay_validation(t *testing.T) {
// 	db := workoutPlanTestDB(t)
// 	plan := models.WorkoutPlan{Name: "Push"}
// 	if err := db.Create(&plan).Error; err != nil {
// 		t.Fatal(err)
// 	}
// 	router := workoutPlanTestRouter(t, db)

// 	cases := []struct {
// 		name       string
// 		body       string
// 		wantStatus int
// 	}{
// 		{"missing field", `{}`, http.StatusBadRequest},
// 		{"null value", `{"day_of_week":null}`, http.StatusBadRequest},
// 		{"too high", `{"day_of_week":7}`, http.StatusBadRequest},
// 		{"negative", `{"day_of_week":-1}`, http.StatusBadRequest},
// 		{"invalid json", `{`, http.StatusBadRequest},
// 		{"saturday ok", `{"day_of_week":6}`, http.StatusOK},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			req := httptest.NewRequest(http.MethodPost, "/workout/plans/1/assign-day", bytes.NewBufferString(tc.body))
// 			req.Header.Set("Content-Type", "application/json")
// 			rec := httptest.NewRecorder()
// 			router.ServeHTTP(rec, req)
// 			if rec.Code != tc.wantStatus {
// 				t.Fatalf("status %d want %d, body: %s", rec.Code, tc.wantStatus, rec.Body.String())
// 			}
// 		})
// 	}
// }

// func TestWorkoutPlanHandler_unassignPlanFromDay(t *testing.T) {
// 	db := workoutPlanTestDB(t)
// 	dow := 3
// 	plan := models.WorkoutPlan{Name: "Pull", DayOfWeek: &dow}
// 	if err := db.Create(&plan).Error; err != nil {
// 		t.Fatal(err)
// 	}
// 	router := workoutPlanTestRouter(t, db)

// 	req := httptest.NewRequest(http.MethodDelete, "/workout/plans/1/assign-day", nil)
// 	rec := httptest.NewRecorder()
// 	router.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusOK {
// 		t.Fatalf("status %d, body: %s", rec.Code, rec.Body.String())
// 	}
// 	var body struct {
// 		Plan struct {
// 			DayOfWeek *int `json:"day_of_week"`
// 		} `json:"plan"`
// 	}
// 	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
// 		t.Fatal(err)
// 	}
// 	if body.Plan.DayOfWeek != nil {
// 		t.Fatalf("expected unassigned day, got %v", body.Plan.DayOfWeek)
// 	}
// }

// func TestWorkoutPlanHandler_getAllWorkoutPlans(t *testing.T) {
// 	db := workoutPlanTestDB(t)
// 	if err := db.Create(&models.WorkoutPlan{Name: "A"}).Error; err != nil {
// 		t.Fatal(err)
// 	}
// 	router := workoutPlanTestRouter(t, db)

// 	req := httptest.NewRequest(http.MethodGet, "/workout/plans/all", nil)
// 	rec := httptest.NewRecorder()
// 	router.ServeHTTP(rec, req)

// 	if rec.Code != http.StatusOK {
// 		t.Fatalf("status %d, body: %s", rec.Code, rec.Body.String())
// 	}
// 	var body struct {
// 		Plans []json.RawMessage `json:"plans"`
// 	}
// 	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(body.Plans) != 1 {
// 		t.Fatalf("plans len %d want 1", len(body.Plans))
// 	}
// }
