package steps

import (
	"net/http"

	"be-simpletracker/internal/core/tracking/common"
	"be-simpletracker/internal/utils/apierr"

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
		apierr.Internal(c, err)
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
		apierr.BadRequest(c, "invalid body")
		return
	}
	date, err := common.ParseDateString(body.Date)
	if err != nil {
		apierr.BadRequest(c, "invalid date")
		return
	}
	row, err := UpsertSteps(c.Request.Context(), date, body.Steps)
	if err != nil {
		common.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": row})
}
