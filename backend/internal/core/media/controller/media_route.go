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
	"be-simpletracker/internal/utils/apierr"

	"github.com/gin-gonic/gin"
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
	group.GET("/catalog", h.catalogSearch)
	group.POST("/catalog/:id/save", h.catalogSave)
	group.PATCH("/catalog/:id", h.catalogUpdate)
	group.DELETE("/catalog/:id", h.catalogDelete)
	group.GET("/:id", h.get)
	group.PATCH("/:id", h.update)
	group.DELETE("/:id", h.delete)
}

// list supports ?type=&status=&q=&favorite=true|false&sort=updated|created|title|rating|finished|release
// and returns the caller's own saves.
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
			apierr.BadRequest(c, "invalid favorite")
			return
		}
		filter.Favorite = &fav
	}
	items, err := service.ListMedia(c.Request.Context(), filter)
	if err != nil {
		respondError(c, err)
		return
	}
	for i := range items {
		items[i].CoverURL = h.safeCover(items[i].CoverURL)
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
	item.CoverURL = h.safeCover(item.CoverURL)
	c.JSON(http.StatusOK, gin.H{"item": item})
}

// create saves a catalog entry for the caller, adding it to the catalog unless one with the same
// source + external_id exists (then that row is reused). 201 for a new save, 200 when the caller
// already had it saved.
func (h *MediaHandler) create(c *gin.Context) {
	var body service.MediaInput
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "invalid request body")
		return
	}
	if !h.coverOK(c, body.CatalogInput) {
		return
	}
	item, created, err := service.CreateMedia(c.Request.Context(), body, clock.Now())
	if err != nil {
		respondError(c, err)
		return
	}
	h.respondSave(c, item, created)
}

func (h *MediaHandler) respondSave(c *gin.Context, item *service.Entry, created bool) {
	item.CoverURL = h.safeCover(item.CoverURL)
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, gin.H{"item": item})
}

// update changes the caller's own fields of a save. Catalog fields go through PATCH /catalog/:id.
func (h *MediaHandler) update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var body service.PersonalInput
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "invalid request body")
		return
	}
	item, err := service.UpdateMedia(c.Request.Context(), id, body, clock.Now())
	if err != nil {
		respondError(c, err)
		return
	}
	item.CoverURL = h.safeCover(item.CoverURL)
	c.JSON(http.StatusOK, gin.H{"item": item})
}

// delete removes the caller's save; the catalog entry stays.
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

// catalogSearch lists shared catalog entries: ?q=&type=
func (h *MediaHandler) catalogSearch(c *gin.Context) {
	items, err := service.SearchCatalog(c.Request.Context(), c.Query("q"), c.Query("type"))
	if err != nil {
		respondError(c, err)
		return
	}
	for i := range items {
		items[i].CoverURL = h.safeCover(items[i].CoverURL)
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// catalogSave saves an existing catalog entry for the caller (optional personal fields in the body).
func (h *MediaHandler) catalogSave(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var body service.PersonalInput
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			apierr.BadRequest(c, "invalid request body")
			return
		}
	}
	item, created, err := service.SaveCatalogItem(c.Request.Context(), id, body, clock.Now())
	if err != nil {
		respondError(c, err)
		return
	}
	h.respondSave(c, item, created)
}

// catalogUpdate edits a catalog entry (its creator or an admin).
func (h *MediaHandler) catalogUpdate(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var body service.CatalogInput
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "invalid request body")
		return
	}
	if !h.coverOK(c, body) {
		return
	}
	item, err := service.UpdateCatalog(c.Request.Context(), id, body)
	if err != nil {
		respondError(c, err)
		return
	}
	item.CoverURL = h.safeCover(item.CoverURL)
	c.JSON(http.StatusOK, gin.H{"item": item})
}

// catalogDelete removes a catalog entry for everyone (admin only).
func (h *MediaHandler) catalogDelete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := service.DeleteCatalog(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// coverOK rejects a cover URL that isn't https on a lookup source's image host (BACKLOG #15): a
// catalog cover loads in every user's browser.
func (h *MediaHandler) coverOK(c *gin.Context, in service.CatalogInput) bool {
	if in.CoverURL.Set && in.CoverURL.Value != nil && !h.lookup.CoverAllowed(*in.CoverURL.Value) {
		apierr.BadRequest(c, "cover_url must be an https image from a supported lookup source")
		return false
	}
	return true
}

// safeCover drops a stored cover that wouldn't pass coverOK (rows from before the rule), so an
// old arbitrary URL is never handed to another user's browser.
func (h *MediaHandler) safeCover(u string) string {
	if h.lookup.CoverAllowed(u) {
		return u
	}
	return ""
}

func parseID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierr.BadRequest(c, "invalid id")
		return 0, false
	}
	return uint(value), true
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalid):
		apierr.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrForbidden):
		apierr.Forbidden(c, "forbidden")
	default:
		apierr.NotFoundOrInternal(c, err)
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
