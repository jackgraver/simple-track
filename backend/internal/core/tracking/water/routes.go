package water

import (
	"errors"
	"net/http"
	"strconv"

	"be-simpletracker/internal/core/tracking/common"
	"be-simpletracker/internal/utils/apierr"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type handler struct{}

func RegisterWaterRoutes(group *gin.RouterGroup) {
	h := handler{}
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
	rows, err := ListWaterLogsForDate(c.Request.Context(), date)
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
	ClientID string  `json:"client_id"`
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
	if body.ClientID != "" {
		if _, err := uuid.Parse(body.ClientID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
			return
		}
	}
	row, err := CreateWaterLog(c.Request.Context(), date, body.AmountOz, body.PresetID, body.ClientID)
	if err != nil {
		// Someone else's preset is ErrPresetNotFound, like a missing one.
		if errors.Is(err, ErrPresetNotFound) || errors.Is(err, ErrAmountNotPositive) {
			apierr.BadRequest(c, err.Error())
			return
		}
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": row})
}

func (h *handler) deleteWater(c *gin.Context) {
	// :id is the numeric id or the entry's client_id (an entry added offline can be deleted before
	// the client knows its server id).
	raw := c.Param("id")
	var err error
	if id64, perr := strconv.ParseUint(raw, 10, 32); perr == nil && id64 != 0 {
		err = DeleteWaterLog(c.Request.Context(), uint(id64))
	} else if u, uerr := uuid.Parse(raw); uerr == nil && u != uuid.Nil {
		err = DeleteWaterLogByClientID(c.Request.Context(), u.String())
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
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

// Preset handlers are the reference for owned endpoints: client errors are 400 with their own
// message, a missing or someone else's preset is the same 404, anything else a logged 500 with a
// fixed message (apierr), never the database's text.

func (h *handler) getPresets(c *gin.Context) {
	rows, err := ListDrinkSizePresets(c.Request.Context())
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"presets": rows})
}

type presetBody struct {
	Name     string  `json:"name"`
	AmountOz float64 `json:"amount_oz"`
}

// presetError answers an error from the preset service.
func presetError(c *gin.Context, err error) {
	if errors.Is(err, ErrPresetNameRequired) || errors.Is(err, ErrAmountNotPositive) {
		apierr.BadRequest(c, err.Error())
		return
	}
	apierr.NotFoundOrInternal(c, err)
}

func (h *handler) postPreset(c *gin.Context) {
	var body presetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "invalid body")
		return
	}
	row, err := CreateDrinkSizePreset(c.Request.Context(), body.Name, body.AmountOz)
	if err != nil {
		presetError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"preset": row})
}

func (h *handler) putPreset(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "invalid id")
		return
	}
	var body presetBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "invalid body")
		return
	}
	row, err := UpdateDrinkSizePreset(c.Request.Context(), uint(id64), body.Name, body.AmountOz)
	if err != nil {
		presetError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"preset": row})
}

func (h *handler) deletePreset(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "invalid id")
		return
	}
	if err := DeleteDrinkSizePreset(c.Request.Context(), uint(id64)); err != nil {
		presetError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
