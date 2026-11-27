package missed

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterMissedRoutes(group *gin.RouterGroup) {
	group.GET("/missed", func(c *gin.Context) {
		date, missingWeight, missingSteps, err := GetMissedYesterday(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"date":   date.Format("2006-01-02"),
			"weight": missingWeight,
			"steps":  missingSteps,
		})
	})
}
