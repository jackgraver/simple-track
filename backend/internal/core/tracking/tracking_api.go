package tracking

import (
	"be-simpletracker/internal/core/tracking/grocery"
	"be-simpletracker/internal/core/tracking/missed"
	"be-simpletracker/internal/core/tracking/profile"
	"be-simpletracker/internal/core/tracking/steps"
	"be-simpletracker/internal/core/tracking/water"
	"be-simpletracker/internal/core/tracking/weight"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware gin.HandlerFunc) {
	group := router.Group("/tracking", authMiddleware)
	missed.RegisterMissedRoutes(group)
	profile.RegisterProfileRoutes(group.Group("/profile"))
	grocery.RegisterGroceryRoutes(group.Group("/grocery"))
	weight.RegisterWeightRoutes(group.Group("/weight"))
	steps.RegisterStepsRoutes(group.Group("/steps"))
	water.RegisterWaterRoutes(group.Group("/water"))
}
