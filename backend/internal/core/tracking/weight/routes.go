package weight

import (
	"net/http"

	"be-simpletracker/internal/core/tracking/common"
	"be-simpletracker/internal/utils/apierr"

	"github.com/gin-gonic/gin"
)

type handler struct{}

func RegisterWeightRoutes(group *gin.RouterGroup) {
	h := handler{}
	group.GET("", h.getWeights)
	group.POST("", h.postWeight)
	group.POST("/import", h.importWeights)
}

func (h *handler) getWeights(c *gin.Context) {
	limit := common.ParseLimitQuery(c)
	rows, err := ListBodyWeights(c.Request.Context(), limit)
	if err != nil {
		apierr.Internal(c, err)
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
		apierr.BadRequest(c, "invalid body")
		return
	}
	date, err := common.ParseDateString(body.Date)
	if err != nil {
		apierr.BadRequest(c, "invalid date")
		return
	}
	row, err := UpsertBodyWeight(c.Request.Context(), date, body.WeightLbs)
	if err != nil {
		common.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": row})
}

type importWeightsBody struct {
	CSV string `json:"csv"`
}

func (h *handler) importWeights(c *gin.Context) {
	var body importWeightsBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "invalid body")
		return
	}
	logs, err := ParseRenphoCSV(body.CSV)
	if err != nil {
		// The parser's messages are ours (row numbers, column names), not the database's.
		apierr.BadRequest(c, err.Error())
		return
	}
	created, updated, err := ImportBodyWeights(c.Request.Context(), logs)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"created": created, "updated": updated})
}
