package grocery

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type handler struct {
	db *gorm.DB
}

func RegisterGroceryRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := handler{db: db}
	group.GET("/items", h.getItems)
	group.POST("/items", h.postItem)
	group.PATCH("/items/:id/complete", h.completeItem)
	group.DELETE("/items/:id", h.deleteItem)
	group.GET("/suggestions", h.getSuggestions)
}

func (h *handler) getItems(c *gin.Context) {
	rows, err := ListActiveItems(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rows})
}

type itemBody struct {
	Name string `json:"name"`
}

func (h *handler) postItem(c *gin.Context) {
	var body itemBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row, err := CreateItem(h.db, body.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": row})
}

func parseIDParam(c *gin.Context) (uint, bool) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return uint(id64), true
}

func (h *handler) completeItem(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	row, err := CompleteItem(h.db, id)
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
	id, ok := parseIDParam(c)
	if !ok {
		return
	}
	if err := DeleteItem(h.db, id); err != nil {
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
	rows, err := ListSuggestions(h.db, c.Query("q"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"suggestions": rows})
}
