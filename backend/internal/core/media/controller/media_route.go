package controller

import (
	"be-simpletracker/internal/clock"
	"errors"
	"log"
	"net/http"
	"strconv"

	"be-simpletracker/internal/core/media/lookup"
	"be-simpletracker/internal/core/media/models"
	"be-simpletracker/internal/core/media/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MediaHandler struct {
	lookup *lookup.Service
}

// RegisterMediaRoutes wires the media routes. The lookup routes never call database.From, so
// they run without a database transaction while they wait on the external sources.
func RegisterMediaRoutes(group *gin.RouterGroup, lookupService *lookup.Service) {
	h := MediaHandler{lookup: lookupService}
	group.GET("", h.list)
	group.POST("", h.create)
	group.GET("/meta", h.meta)
	group.GET("/stats", h.stats)
	group.GET("/lookup", h.lookupSearch)
	group.GET("/lookup/details", h.lookupDetails)
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
	items, err := service.ListMedia(c.Request.Context(), filter)
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
	stats, err := service.GetMediaStats(c.Request.Context(), clock.Now())
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
	item, err := service.GetMedia(c.Request.Context(), id)
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
	item, err := service.CreateMedia(c.Request.Context(), body, clock.Now())
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
	item, err := service.UpdateMedia(c.Request.Context(), id, body, clock.Now())
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
	if err := service.DeleteMedia(c.Request.Context(), id); err != nil {
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

// lookupSearch returns candidate matches: ?type=movie&q=heat
func (h *MediaHandler) lookupSearch(c *gin.Context) {
	results, err := h.lookup.Search(c.Request.Context(), models.MediaType(c.Query("type")), c.Query("q"))
	if err != nil {
		respondLookupError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

// lookupDetails returns the full record for one match: ?source=tmdb&id=movie/949
func (h *MediaHandler) lookupDetails(c *gin.Context) {
	result, err := h.lookup.Details(c.Request.Context(), c.Query("source"), c.Query("id"))
	if err != nil {
		respondLookupError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": result})
}

// respondLookupError maps lookup errors to fixed client messages. The error text can
// describe the upstream request (URL, API key, response body), so it is only logged
// (redacted) and never sent to the client.
func respondLookupError(c *gin.Context, err error) {
	status, msg := http.StatusBadGateway, "Lookup service unavailable"
	switch {
	case errors.Is(err, lookup.ErrNotConfigured):
		status, msg = http.StatusServiceUnavailable, "Lookup is not configured"
	case errors.Is(err, lookup.ErrUnsupported):
		status, msg = http.StatusBadRequest, "Lookup is not supported for this media type"
	case errors.Is(err, lookup.ErrNotFound):
		status, msg = http.StatusNotFound, "Lookup result not found"
	}
	if status != http.StatusNotFound && status != http.StatusBadRequest {
		log.Printf("media lookup failed (%s %s): %s", c.Request.Method, c.Request.URL.Path, lookup.Detail(err))
	}
	c.JSON(status, gin.H{"error": msg})
}
