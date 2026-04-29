package routes

import (
	"be-simpletracker/internal/core/diet/services"
	"be-simpletracker/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DietPlanHandler struct {
	db *gorm.DB
}

func NewDietPlanHandler(db *gorm.DB) *DietPlanHandler {
	return &DietPlanHandler{db: db}
}

func RegisterDietPlanRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewDietPlanHandler(db)
	plans := group.Group("/plans")
	{
		plans.GET("/plan/all", h.getAllPlans)
	}
}

func (h *DietPlanHandler) getAllPlans(c *gin.Context) {
	ctx := c.Request.Context()
	params := utils.ParseQueryParams(c)
	result, err := services.GetAllPlans(ctx, h.db, params)
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
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"plans": &result.Data,
	})
}
