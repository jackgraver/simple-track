package profile

import (
	"net/http"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := UpsertProfile(c.Request.Context(), body.HeightIn, body.Age, body.Sex, body.ActivityLevel)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": row})
}
