package profile

import (
	"net/http"

	"be-simpletracker/internal/core/tracking/common"
	"be-simpletracker/internal/utils/apierr"

	"github.com/gin-gonic/gin"
)

type handler struct{}

func RegisterProfileRoutes(group *gin.RouterGroup) {
	h := handler{}
	group.GET("", h.getProfile)
	group.PUT("", h.putProfile)
}

func (h *handler) getProfile(c *gin.Context) {
	row, err := GetProfile(c.Request.Context())
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if row == nil {
		c.JSON(http.StatusOK, gin.H{"profile": nil})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": row})
}

type putProfileBody struct {
	HeightIn      float64 `json:"height_in"`
	Age           int     `json:"age"`
	Sex           string  `json:"sex"`
	ActivityLevel string  `json:"activity_level"`
}

func (h *handler) putProfile(c *gin.Context) {
	var body putProfileBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "invalid body")
		return
	}
	row, err := UpsertProfile(c.Request.Context(), body.HeightIn, body.Age, body.Sex, body.ActivityLevel)
	if err != nil {
		common.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": row})
}
