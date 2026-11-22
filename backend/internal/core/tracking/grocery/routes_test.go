package grocery

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func groceryRouter(t *testing.T) (*gin.Engine, func(string) *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := groceryTestDB(t)
	r := gin.New()
	RegisterGroceryRoutes(r.Group("/grocery"), db)
	post := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/grocery/items", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	t.Cleanup(func() {
		if d, err := db.DB(); err == nil {
			d.Close()
		}
	})
	return r, post
}

// Review finding 4: client_id is validated up front, and database errors are not echoed.
func TestPostItem_validatesClientIDAndHidesDBErrors(t *testing.T) {
	_, post := groceryRouter(t)
	if w := post(`{"name": "milk", "client_id": "nope"}`); w.Code != http.StatusBadRequest ||
		!strings.Contains(w.Body.String(), "invalid client_id") {
		t.Fatalf("bad client_id: %d %s", w.Code, w.Body.String())
	}
	if w := post(`{"name": "  "}`); w.Code != http.StatusBadRequest {
		t.Fatalf("empty name: %d %s", w.Code, w.Body.String())
	}
	if w := post(`{"name": "milk", "client_id": "55555555-5555-4555-8555-555555555555"}`); w.Code != http.StatusOK {
		t.Fatalf("valid create: %d %s", w.Code, w.Body.String())
	}
}

func TestPostItem_dbErrorIs500WithoutDetails(t *testing.T) {
	db := groceryTestDB(t)
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterGroceryRoutes(r.Group("/grocery"), db)
	if err := db.Exec("DROP TABLE grocery_items").Error; err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/grocery/items", strings.NewReader(`{"name": "milk"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), "grocery_items") {
		t.Fatalf("want opaque 500, got %d %s", w.Code, w.Body.String())
	}
}

// A row created offline is completed by its client_id before the client knows the server id.
func TestCompleteItem_byClientID(t *testing.T) {
	r, post := groceryRouter(t)
	const cid = "66666666-6666-4666-8666-666666666666"
	if w := post(`{"name": "eggs", "client_id": "` + cid + `"}`); w.Code != http.StatusOK {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	patch := func(id string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPatch, "/grocery/items/"+id+"/complete", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	for i := 0; i < 2; i++ { // idempotent replay
		if w := patch(cid); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"completed_at":"`) {
			t.Fatalf("complete by client_id (%d): %d %s", i, w.Code, w.Body.String())
		}
	}
	if w := patch("77777777-7777-4777-8777-777777777777"); w.Code != http.StatusNotFound {
		t.Fatalf("unknown client_id: %d %s", w.Code, w.Body.String())
	}
	if w := patch("nope"); w.Code != http.StatusBadRequest {
		t.Fatalf("bad id: %d %s", w.Code, w.Body.String())
	}
}
