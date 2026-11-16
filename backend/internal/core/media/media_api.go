package media

import (
	"be-simpletracker/internal/core/media/controller"
	"be-simpletracker/internal/core/media/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Migrate() error {
	return h.db.AutoMigrate(&models.MediaItem{})
}

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	controller.RegisterMediaRoutes(router.Group("/media", authMiddleware), h.db)
}
