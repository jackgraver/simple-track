//go:build e2e

// Package e2e is the test-only backend surface for the Playwright suite (e2e/README.md), compiled
// only with `-tags e2e`. It adds, under /__e2e (no auth):
//
//	POST   /__e2e/reset                 truncate + reseed, clear clock, faults and holds
//	POST   /__e2e/seed/:profile         add a data profile on top of the seed (e.g. "large")
//	POST   /__e2e/clock  {time}         move the server clock (RFC 3339); DELETE resets it
//	POST   /__e2e/faults {table,op,...} make matching DB statements fail
//	POST   /__e2e/holds  {method,path,phase,...} hold matching requests
//	GET    /__e2e/holds/:id?wait=ms     long-poll until the held request arrives
//	POST   /__e2e/holds/:id/{release,fail,drop}
//	GET    /__e2e/state                 faults and holds with hit counts
//
// The server refuses to start this build in production.
package e2e

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/e2e/seed"
	"be-simpletracker/internal/env"
)

// Enabled reports whether this binary was built with the e2e tag.
const Enabled = true

// Install wires the hold middleware, the DB fault hooks and the /__e2e routes. Call it before
// registering the app's routes: gin applies middleware only to routes added after Use.
func Install(router *gin.Engine, db *gorm.DB) {
	if env.IsProduction() {
		log.Fatal("e2e: this binary was built with -tags e2e and must not run in production")
	}
	log.Print("e2e: test-only endpoints enabled under /__e2e")
	s := newState(router, db)
	// Reset and seed profiles truncate and reseed as the schema owner, like cmd/e2eseed. The
	// server's own pool is the app role, which may not TRUNCATE or restart sequences.
	owner, err := database.OwnerDSN()
	if err != nil {
		log.Fatalf("e2e: owner database: %v", err)
	}
	if s.seedDB, err = database.Open(owner, database.Pool{MaxOpen: 2, MaxIdle: 1, MaxLifetime: time.Hour, MaxIdleTime: time.Hour}); err != nil {
		log.Fatalf("e2e: owner database: %v", err)
	}
	if err := s.registerFaultHooks(db); err != nil {
		log.Fatalf("e2e: %v", err)
	}
	router.Use(s.holdMiddleware())
	s.routes(router.Group("/__e2e"))
}

func (s *state) routes(g *gin.RouterGroup) {
	g.POST("/reset", func(c *gin.Context) {
		s.clear()
		clock.Reset()
		if err := seed.Reset(s.seedDB); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	g.POST("/seed/:profile", func(c *gin.Context) {
		apply, ok := seed.Profiles[c.Param("profile")]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "unknown seed profile " + c.Param("profile")})
			return
		}
		if err := apply(s.seedDB); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	g.POST("/clock", func(c *gin.Context) {
		var body struct {
			Time string `json:"time" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		t, err := time.Parse(time.RFC3339Nano, body.Time)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "time must be RFC 3339"})
			return
		}
		clock.Set(t)
		c.JSON(http.StatusOK, gin.H{"now": clock.Now()})
	})
	g.DELETE("/clock", func(c *gin.Context) {
		clock.Reset()
		c.Status(http.StatusNoContent)
	})

	g.POST("/faults", func(c *gin.Context) {
		var f fault
		if err := c.ShouldBindJSON(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := s.addFault(f)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id})
	})

	g.POST("/holds", func(c *gin.Context) {
		var h holdSpec
		if err := c.ShouldBindJSON(&h); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		id, err := s.addHold(h)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id})
	})
	g.GET("/holds/:id", func(c *gin.Context) {
		wait, _ := time.ParseDuration(c.DefaultQuery("wait", "0") + "ms")
		arrived, err := s.waitArrived(c.Param("id"), wait)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"arrived": arrived})
	})
	g.POST("/holds/:id/:action", func(c *gin.Context) {
		d := decision{action: c.Param("action")}
		if d.action == "fail" {
			var f fault
			if err := c.ShouldBindJSON(&f); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			d.fault = &f
		}
		if err := s.decide(c.Param("id"), d); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	g.GET("/state", func(c *gin.Context) {
		c.JSON(http.StatusOK, s.snapshot())
	})
}

func isE2EPath(p string) bool { return strings.HasPrefix(p, "/__e2e/") }
