package routes

import (
	"be-simpletracker/features/diet/models"
	"be-simpletracker/generics"
	"be-simpletracker/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DietPlanHandler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewDietPlanHandler(db *gorm.DB) *DietPlanHandler {
	return &DietPlanHandler{db: db}
}

func RegisterDietPlanRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewDietPlanHandler(db)

	config := generics.DefaultCRUDConfig[models.Plan]("/plans", "plan")
	generics.RegisterBasicCRUD(group, db, config)

	plans := group.Group("/plans")
	{
		plans.GET("/plan/all", h.getAllPlans)
	}
}

func (h *DietPlanHandler) getAllPlans(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Use the service function - encapsulates all repository logic
	// result, err := services.GetAllPlans(ctx, f.db, c)
    result, err := utils.GetAllWithOptions[*models.Plan](ctx, h.db, c, "id", true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.Pagination != nil {
		c.JSON(http.StatusOK, gin.H{
			"plans": &result.Data,
			"pagination": gin.H{
				"total":      result.Pagination.Total,
				"page":       result.Pagination.Page,
				"pageSize":   result.Pagination.PageSize,
				"totalPages": result.Pagination.TotalPages,
				"hasNext":    result.Pagination.HasNext,
				"hasPrev":    result.Pagination.HasPrev,
			},
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"plans": &result.Data,
		})
	}
}