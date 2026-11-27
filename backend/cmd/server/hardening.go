package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// defaultBodyLimit caps request bodies on every route not listed in bodyLimits (BACKLOG #8).
// API requests are small JSON documents; Caddy's request_body max_size (deploy/Caddyfile) is the
// outer guard and must stay at or above the largest limit here.
const defaultBodyLimit int64 = 1 << 20 // 1 MiB

// bodyLimits are routes allowed a larger body, keyed by "METHOD /gin/full/path".
// TestBodyLimitRoutesExist fails if a key stops matching a registered route.
var bodyLimits = map[string]int64{
	// Renpho CSV export sent as a JSON string; also capped by row count in weight.ParseRenphoCSV.
	"POST /tracking/weight/import": 5 << 20,
}

// hardening applies to every response: Cache-Control: no-store (BACKLOG #17), since API responses
// carry personal data and must not sit in browser, bfcache or intermediary caches, and the
// request body limit (#8). It runs after routing, so c.FullPath() is known.
func hardening(def int64, limits map[string]int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		limit := def
		if n, ok := limits[c.Request.Method+" "+c.FullPath()]; ok {
			limit = n
		}
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		// Chunked or under-reported bodies fail on read instead, so binding returns an error.
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
