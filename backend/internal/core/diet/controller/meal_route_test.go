package controller_test

import (
	"be-simpletracker/internal/database"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"be-simpletracker/internal/core/diet/controller"
	"be-simpletracker/internal/core/diet/testutil"

	"github.com/gin-gonic/gin"
)

func dietRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := testutil.SetupTestDB(t)
	r := gin.New()
	r.Use(database.Middleware(db))
	r.POST("/quick-log", controller.PostQuickLog)
	r.POST("/planned/from-saved", controller.PostPlannedFromSaved)
	r.DELETE("/meal/logged", controller.DeleteLoggedMeal)
	r.POST("/foods", controller.PostFood)
	return r
}

func send(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Review finding 4: a malformed client_id is a 400, not a uuid-cast 500 from Postgres.
func TestDietRoutes_rejectNonUUIDClientID(t *testing.T) {
	r := dietRouter(t)
	for _, tc := range []struct{ path, body string }{
		{"/quick-log", `{"name": "Bar", "calories": 100, "client_id": "not-a-uuid"}`},
		{"/planned/from-saved", `{"saved_meal_id": 1, "client_id": "not-a-uuid"}`},
	} {
		w := send(r, http.MethodPost, tc.path, tc.body)
		if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "client_id") {
			t.Errorf("%s: want 400 about client_id, got %d %s", tc.path, w.Code, w.Body.String())
		}
	}
}

// Review finding 6: unlogging from a day that doesn't exist is a 404.
func TestDeleteLoggedMeal_unknownDayIs404(t *testing.T) {
	r := dietRouter(t)
	w := send(r, http.MethodDelete, "/meal/logged", `{"day_id": 9999, "meal_id": 1}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d %s", w.Code, w.Body.String())
	}
}

// e2e backlog: a duplicate food name is a 409 with a readable message, not a 500.
func TestPostFood_duplicateNameIs409(t *testing.T) {
	r := dietRouter(t)
	body := `{"name": "Oats", "serving_type": "g", "serving_amount": 40, "calories": 150}`
	if w := send(r, http.MethodPost, "/foods", body); w.Code != http.StatusCreated {
		t.Fatalf("first create: want 201, got %d %s", w.Code, w.Body.String())
	}
	w := send(r, http.MethodPost, "/foods", body)
	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "already exists") {
		t.Fatalf("duplicate: want 409, got %d %s", w.Code, w.Body.String())
	}
}
