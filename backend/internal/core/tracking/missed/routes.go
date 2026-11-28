package missed

import (
	"net/http"

	"be-simpletracker/internal/utils/apierr"

	"github.com/gin-gonic/gin"
)

func RegisterMissedRoutes(group *gin.RouterGroup) {
	group.GET("/missed", func(c *gin.Context) {
		date, missingWeight, missingSteps, err := GetMissedYesterday(c.Request.Context())
		if err != nil {
			apierr.Internal(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"date":   date.Format("2006-01-02"),
			"weight": missingWeight,
			"steps":  missingSteps,
		})
	})
}
