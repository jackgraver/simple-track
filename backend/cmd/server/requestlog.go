package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// requestLogger replaces gin's default logger: it logs the matched route pattern
// (/media/:id), never the raw path or query string, so search terms and anything else users put
// in URLs stay out of the server log (BACKLOG #19). Unmatched paths log as "(no route)".
func requestLogger() []gin.HandlerFunc {
	log := gin.LoggerWithFormatter(func(p gin.LogFormatterParams) string {
		route, _ := p.Keys[routeKey].(string)
		if route == "" {
			route = "(no route)"
		}
		return fmt.Sprintf("[GIN] %s | %3d | %10v | %15s | %-7s %s\n",
			p.TimeStamp.Format("2006/01/02 - 15:04:05"), p.StatusCode,
			p.Latency.Round(time.Microsecond), p.ClientIP, p.Method, route)
	})
	// LogFormatterParams has no FullPath, so it travels in the context keys.
	record := func(c *gin.Context) {
		c.Set(routeKey, c.FullPath())
		c.Next()
	}
	return []gin.HandlerFunc{log, record}
}

const routeKey = "requestlog.route"
