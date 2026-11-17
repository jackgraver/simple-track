package money

import (
	"be-simpletracker/internal/core/money/controller"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	controller.RegisterRoutes(router.Group("/money", authMiddleware), h.db)
}
