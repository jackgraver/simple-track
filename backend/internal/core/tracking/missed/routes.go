package missed

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterMissedRoutes(group *gin.RouterGroup, db *gorm.DB) {
	group.GET("/missed", func(c *gin.Context) {
		date, missingWeight, missingSteps, err := GetMissedYesterday(db)
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
