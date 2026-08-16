package controller

import (
	"be-simpletracker/internal/core/diet/services"
	"be-simpletracker/internal/utils"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var planQueryPolicy = utils.QueryPolicy{
	SortableFields: map[string]string{
		"id":             "id",
		"name":           "name",
		"created_at":     "created_at",
		"updated_at":     "updated_at",
		"effective_from": "effective_from",
	},
	FilterableFields: map[string]string{
		"id":             "id",
		"name":           "name",
		"effective_from": "effective_from",
	},
}

type updatePlanMacrosRequest struct {
	Calories float32 `json:"calories"`
	Protein  float32 `json:"protein"`
	Fiber    float32 `json:"fiber"`
	Carbs    float32 `json:"carbs"`
	Fat      float32 `json:"fat"`
}

func GetAllPlans(c *gin.Context) {
	ctx := c.Request.Context()
	params, err := utils.ParseQueryParams(c, planQueryPolicy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := services.GetAllPlans(ctx, params)
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

func PutPlanMacros(c *gin.Context) {
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
	if req.Calories < 0 || req.Protein < 0 || req.Fiber < 0 || req.Carbs < 0 || req.Fat < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "macro targets must be non-negative"})
		return
	}
	plan, err := services.UpdatePlanMacros(uint(id64), req.Calories, req.Protein, req.Fiber, req.Carbs, req.Fat)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"plan": plan})
}
