package water

import (
	"errors"
	"net/http"
	"strconv"

	"be-simpletracker/internal/core/tracking/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	db *gorm.DB
}

func RegisterWaterRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := handler{db: db}
	group.GET("", h.getWater)
	group.POST("", h.postWater)
	group.DELETE("/:id", h.deleteWater)
	p := group.Group("/presets")
	{
		p.GET("", h.getPresets)
		p.POST("", h.postPreset)
		p.PUT("/:id", h.putPreset)
		p.DELETE("/:id", h.deletePreset)
	}
}

func (h *handler) getWater(c *gin.Context) {
	dateStr := c.Query("date")
	date, err := common.ParseDateString(dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	rows, err := ListWaterLogsForDate(h.db, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"logs": rows})
}

type postWaterBody struct {
	Date     string  `json:"date"`
	AmountOz float64 `json:"amount_oz"`
	PresetID *uint   `json:"preset_id"`
}

func (h *handler) postWater(c *gin.Context) {
	var body postWaterBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := common.ParseDateString(body.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date"})
		return
	}
	row, err := CreateWaterLog(h.db, date, body.AmountOz, body.PresetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": row})
}

func (h *handler) deleteWater(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = DeleteWaterLog(h.db, uint(id64))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *handler) getPresets(c *gin.Context) {
	rows, err := ListDrinkSizePresets(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"presets": rows})
}

type presetBody struct {
	Name     string  `json:"name"`
	AmountOz float64 `json:"amount_oz"`
}

func (h *handler) postPreset(c *gin.Context) {
	var body presetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := CreateDrinkSizePreset(h.db, body.Name, body.AmountOz)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"preset": row})
}

func (h *handler) putPreset(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body presetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := UpdateDrinkSizePreset(h.db, uint(id64), body.Name, body.AmountOz)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"preset": row})
}

func (h *handler) deletePreset(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	err = DeleteDrinkSizePreset(h.db, uint(id64))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
