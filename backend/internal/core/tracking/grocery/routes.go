package grocery

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"be-simpletracker/internal/utils/apierr"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type handler struct{}

func RegisterGroceryRoutes(group *gin.RouterGroup) {
	h := handler{}
	group.GET("/items", h.getItems)
	group.POST("/items", h.postItem)
	group.POST("/items/parse", h.parseItems)
	group.POST("/items/bulk", h.postItems)
	group.PATCH("/items/:id/complete", h.completeItem)
	group.DELETE("/items/:id", h.deleteItem)
	group.GET("/suggestions", h.getSuggestions)
}

func (h *handler) getItems(c *gin.Context) {
	rows, err := ListActiveItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

type itemBody struct {
	Name     string `json:"name"`
	ClientID string `json:"client_id"`
}

func (h *handler) postItem(c *gin.Context) {
	var body itemBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	clientID := strings.TrimSpace(body.ClientID)
	if clientID != "" {
		id, err := uuid.Parse(clientID)
		if err != nil || id == uuid.Nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid client_id"})
			return
		}
		clientID = id.String()
	}
	row, err := CreateItem(c.Request.Context(), body.Name, clientID)
	if err != nil {
		if errors.Is(err, ErrNameRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Don't echo driver errors to the client.
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": row})
}

// itemRef is the `:id` path param: a numeric id, or the client_id of a row the client created
// (so a row added offline can be completed before the client knows its server id).
type itemRef struct {
	id       uint
	clientID string
}

func parseItemRef(c *gin.Context) (itemRef, bool) {
	raw := c.Param("id")
	if id64, err := strconv.ParseUint(raw, 10, 32); err == nil && id64 != 0 {
		return itemRef{id: uint(id64)}, true
	}
	if u, err := uuid.Parse(raw); err == nil && u != uuid.Nil {
		return itemRef{clientID: u.String()}, true
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
	return itemRef{}, false
}

func (h *handler) completeItem(c *gin.Context) {
	ref, ok := parseItemRef(c)
	if !ok {
		return
	}
	row, err := CompleteItem(c.Request.Context(), ref)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": row})
}

func (h *handler) deleteItem(c *gin.Context) {
	ref, ok := parseItemRef(c)
	if !ok {
		return
	}
	if err := DeleteItem(c.Request.Context(), ref); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *handler) getSuggestions(c *gin.Context) {
	rows, err := ListSuggestions(c.Request.Context(), c.Query("q"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"suggestions": rows})
}

type parseBody struct {
	Text string `json:"text"`
}

// POST /items/parse — split pasted text into items matched to known names.
func (h *handler) parseItems(c *gin.Context) {
	var body parseBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	lines, err := ParseTextList(c.Request.Context(), body.Text)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"lines": lines})
}

type bulkBody struct {
	Names []string `json:"names"`
}

// POST /items/bulk — add several items, skipping ones already on the list.
func (h *handler) postItems(c *gin.Context) {
	var body bulkBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, skipped, err := CreateItems(c.Request.Context(), body.Names)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": created, "skipped": skipped})
}
