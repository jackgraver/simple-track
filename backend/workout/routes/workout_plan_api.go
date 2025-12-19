package routes

import (
	"be-simpletracker/database/repository"
	generics "be-simpletracker/generics"
	"be-simpletracker/workout/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WorkoutPlanHandler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewWorkoutPlanHandler(db *gorm.DB) *WorkoutPlanHandler {
	return &WorkoutPlanHandler{db: db}
}

// RegisterPlanRoutes registers all plan routes under the workout group
func RegisterWorkoutPlanRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewWorkoutPlanHandler(db)

	plans := group.Group("/plans")
	{
		plans.GET("/all", h.GetAll)
		plans.GET("/:id", h.GetByID)
		plans.POST("", h.Create)
		plans.PUT("/:id", h.Update)
		plans.DELETE("/:id", h.Delete)

		// group.POST("/plans/:id/exercises/add", f.addExerciseToPlan)
   		// group.DELETE("/plans/:id/exercises/remove", f.removeExerciseFromPlan)
	}
}

// GetAll retrieves all workout plans with support for pagination, sorting, and filtering
func (h *WorkoutPlanHandler) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Parse query parameters for pagination
	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "0")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	
	// Build query options
	var opts []repository.QueryOption
	opts = append(opts, repository.WithDefaultPreloads()) // Preload Exercises
	
	// Add sorting
	orderBy := c.Query("orderBy")
	if orderBy != "" {
		orderDesc := c.DefaultQuery("orderDesc", "false") == "true"
		opts = append(opts, repository.WithOrderBy(orderBy, orderDesc))
	} else {
		opts = append(opts, repository.WithOrderByDesc("id"))
	}
	
	// Add filters
	if nameFilter := c.Query("name"); nameFilter != "" {
		opts = append(opts, repository.WithFilter("name", nameFilter))
	}
	
	// Execute query with pagination or without
	if page > 0 && pageSize > 0 {
		result, err := generics.GetAllPaginated[models.WorkoutPlan](ctx, h.db, page, pageSize, opts...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"plans": result.Data,
			"pagination": gin.H{
				"total":      result.Total,
				"page":       result.Page,
				"pageSize":   result.PageSize,
				"totalPages": result.TotalPages,
				"hasNext":    result.HasNext,
				"hasPrev":    result.HasPrev,
			},
		})
	} else {
		// Use generic GetAll - WorkoutPlan has default preloads for Exercises
		plans, err := generics.GetAll[models.WorkoutPlan](ctx, h.db, opts...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"plans": plans,
		})
	}
}

// GetByID retrieves a single workout plan by ID
func (h *WorkoutPlanHandler) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Use generic GetOne (or GetByID) with default preloads (Exercises)
	plan, err := generics.GetOne[models.WorkoutPlan](ctx, h.db, uint(id), repository.WithDefaultPreloads())
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

// Create creates a new workout plan
func (h *WorkoutPlanHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	
	var plan models.WorkoutPlan
	if err := c.BindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use generic Create
	if err := generics.Create(ctx, h.db, plan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"plan": plan})
}

// Update updates an existing workout plan
func (h *WorkoutPlanHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var plan models.WorkoutPlan
	if err := c.BindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure ID matches URL parameter
	plan.ID = uint(id)

	// Use generic Update
	if err := generics.Update(ctx, h.db, plan); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"plan": plan})
}

// Delete deletes a workout plan by ID
func (h *WorkoutPlanHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Use generic Delete
	if err := generics.Delete[models.WorkoutPlan](ctx, h.db, uint(id)); err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
