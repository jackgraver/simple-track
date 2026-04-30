package weight

import (
	"net/http"

	"be-simpletracker/internal/core/tracking/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	db *gorm.DB
}

func RegisterWeightRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := handler{db: db}
	group.GET("", h.getWeights)
	group.POST("", h.postWeight)
}

func (h *handler) getWeights(c *gin.Context) {
	limit := common.ParseLimitQuery(c)
	rows, err := ListBodyWeights(h.db, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": rows})
}

type postWeightBody struct {
	Date      string  `json:"date"`
	WeightLbs float64 `json:"weight_lbs"`
}

func (h *handler) postWeight(c *gin.Context) {
	var body postWeightBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := common.ParseDateString(body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	row, err := UpsertBodyWeight(h.db, date, body.WeightLbs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": row})
}
