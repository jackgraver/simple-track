package routes

import (
	"be-simpletracker/internal/core/diet/models"
	"be-simpletracker/internal/core/diet/services"
	"be-simpletracker/internal/utils"
	"errors"
	"net/http"
	"strconv"

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
		plans.PUT("/plan/:id", h.putPlanMacros)
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

type updatePlanMacrosRequest struct {
	Calories float32 `json:"calories"`
	Protein  float32 `json:"protein"`
	Fiber    float32 `json:"fiber"`
	Carbs    float32 `json:"carbs"`
}

func (h *DietPlanHandler) putPlanMacros(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid plan id"})
		return
	}
	var req updatePlanMacrosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Calories < 0 || req.Protein < 0 || req.Fiber < 0 || req.Carbs < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "macro targets must be non-negative"})
		return
	}
	var plan models.Plan
	if err := h.db.First(&plan, uint(id64)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	plan.Calories = req.Calories
	plan.Protein = req.Protein
	plan.Fiber = req.Fiber
	plan.Carbs = req.Carbs
	if err := h.db.Model(&plan).Updates(map[string]any{
		"calories": req.Calories,
		"protein":  req.Protein,
		"fiber":    req.Fiber,
		"carbs":    req.Carbs,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": plan})
}
