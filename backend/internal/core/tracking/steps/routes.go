package steps

import (
	"net/http"

	"be-simpletracker/internal/core/tracking/common"

	"github.com/gin-gonic/gin"
)

type handler struct{}

func RegisterStepsRoutes(group *gin.RouterGroup) {
	h := handler{}
	group.GET("", h.getSteps)
	group.POST("", h.postSteps)
}

func (h *handler) getSteps(c *gin.Context) {
	limit := common.ParseLimitQuery(c)
	rows, err := ListSteps(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": rows})
}

type postStepsBody struct {
	Date  string `json:"date"`
	Steps int    `json:"steps"`
}

func (h *handler) postSteps(c *gin.Context) {
	var body postStepsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := common.ParseDateString(body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	row, err := UpsertSteps(c.Request.Context(), date, body.Steps)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": row})
}
