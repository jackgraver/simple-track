package money

import (
	"be-simpletracker/internal/core/money/controller"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	controller.RegisterRoutes(router.Group("/money", authMiddleware))
}
