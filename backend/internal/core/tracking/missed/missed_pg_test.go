package missed

import (
	"net/http"
	"testing"

	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/tracking/steps"
	"be-simpletracker/internal/core/tracking/weight"
	"be-simpletracker/internal/database/twousers"

	"github.com/gin-gonic/gin"
)

// Yesterday's gaps are per user: Alice logging doesn't hide Bob's missing entries.
func TestMissed_isPerUser(t *testing.T) {
	h := twousers.New(t)
	api := h.Router(func(g *gin.RouterGroup) {
		RegisterMissedRoutes(g)
		weight.RegisterWeightRoutes(g.Group("/weight"))
		steps.RegisterStepsRoutes(g.Group("/steps"))
	})
	yesterday := clock.ZerodTime(1).Format("2006-01-02")
	api.Do(h.Alice, http.MethodPost, "/weight", map[string]any{"date": yesterday, "weight_lbs": 180}).Status(t, http.StatusOK)
	api.Do(h.Alice, http.MethodPost, "/steps", map[string]any{"date": yesterday, "steps": 5000}).Status(t, http.StatusOK)

	missed := func(u twousers.User) (weight, steps bool) {
		var out struct{ Weight, Steps bool }
		api.Do(u, http.MethodGet, "/missed", nil).Status(t, http.StatusOK).JSON(t, &out)
		return out.Weight, out.Steps
	}
	if w, s := missed(h.Alice); w || s {
		t.Fatalf("Alice logged both but missed = weight %v steps %v", w, s)
	}
	if w, s := missed(h.Bob); !w || !s {
		t.Fatalf("Bob logged nothing but missed = weight %v steps %v, want both", w, s)
	}
}
