package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// healthHandler reports ok only when Postgres answers. It is registered
// after every migration has run, and the server only starts listening
// after that, so a 200 also means the schema is up to date. The deploy
// agent's health check (the compose healthcheck on /health) relies on
// both: a release that can't reach or migrate its database never counts
// as healthy, and gets rolled back.
func healthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := pingDB(c.Request.Context(), db); err != nil {
			// Details go to the log, not the response: /health is public.
			log.Printf("health: database unreachable: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

func pingDB(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return sqlDB.PingContext(ctx)
}
