package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"be-simpletracker/internal/core/media/models"
	"be-simpletracker/internal/core/media/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MediaHandler struct {
	db *gorm.DB
}

func RegisterMediaRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := MediaHandler{db: db}
	group.GET("", h.list)
	group.POST("", h.create)
	group.GET("/meta", h.meta)
	group.GET("/stats", h.stats)
	group.GET("/:id", h.get)
	group.PATCH("/:id", h.update)
	group.DELETE("/:id", h.delete)
}

// list supports ?type=&status=&q=&favorite=true|false&sort=updated|created|title|rating|finished|release
func (h *MediaHandler) list(c *gin.Context) {
	filter := service.ListFilter{
		Type:   c.Query("type"),
		Status: c.Query("status"),
		Query:  c.Query("q"),
		Sort:   c.Query("sort"),
	}
	if raw := c.Query("favorite"); raw != "" {
		fav, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid favorite"})
			return
		}
		filter.Favorite = &fav
	}
	items, err := service.ListMedia(h.db, filter)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *MediaHandler) meta(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"types": models.MediaTypes, "statuses": models.MediaStatuses})
}

func (h *MediaHandler) stats(c *gin.Context) {
	stats, err := service.GetMediaStats(h.db, time.Now())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (h *MediaHandler) get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := service.GetMedia(h.db, id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *MediaHandler) create(c *gin.Context) {
	var body service.MediaInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := service.CreateMedia(h.db, body, time.Now())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"item": item})
}

func (h *MediaHandler) update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var body service.MediaInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := service.UpdateMedia(h.db, id, body, time.Now())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": item})
}

func (h *MediaHandler) delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := service.DeleteMedia(h.db, id); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func parseID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(value), true
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "media item not found"})
	case errors.Is(err, service.ErrInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
