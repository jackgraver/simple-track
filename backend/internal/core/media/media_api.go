package media

import (
	"be-simpletracker/internal/core/media/controller"
	"be-simpletracker/internal/core/media/lookup"
	"be-simpletracker/internal/env"

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
	lookupService := lookup.NewService(env.OptionalString("TMDB_API_KEY"))
	controller.RegisterMediaRoutes(router.Group("/media", authMiddleware), h.db, lookupService)
}
