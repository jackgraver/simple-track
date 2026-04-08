package routes

import (
	"be-simpletracker/internal/core/diet/models"
	"be-simpletracker/internal/core/diet/services"
	"be-simpletracker/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MealHandler struct {
	db *gorm.DB
}

func NewMealHandler(db *gorm.DB) *MealHandler {
	return &MealHandler{db: db}
}

func RegisterMealRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewMealHandler(db)
	foods := group.Group("/foods")
	{
		foods.POST("", h.postFood)
	}
	meals := group.Group("/meals")
	{
		meals.GET("/food/all", h.getAllFoods)
		meals.GET("/meal/all", h.getAllMeals)
		meals.GET("/meal/:id", h.getMeal)
		meals.POST("/meal/new", h.postNewMeal)
		meals.POST("/meal/log-planned", h.postLogPlanned)
		meals.POST("/meal/logedited", h.postLogEdited)
		meals.POST("/meal/editlogged", h.postEditLogged)
		meals.DELETE("/meal/logged", h.deleteLoggedMeal)
	}
}

func (h *MealHandler) postFood(c *gin.Context) {
	var food models.Food
	if err := c.ShouldBindJSON(&food); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdFood, err := services.CreateFood(h.db, &food)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"food": createdFood})
}

func (h *MealHandler) getAllFoods(c *gin.Context) {
	excludeIDsStr := c.Query("exclude")
	var excludeIDs []uint
	if excludeIDsStr != "" {
		if id, err := strconv.ParseUint(excludeIDsStr, 10, 32); err == nil {
			excludeIDs = append(excludeIDs, uint(id))
		}
	}
	foods, err := services.AllFoods(h.db, excludeIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"foods": foods})
}

func (h *MealHandler) getAllMeals(c *gin.Context) {
	excludeIDsStr := c.Query("exclude")
	var excludeIDs []uint
	if excludeIDsStr != "" {
		if id, err := strconv.ParseUint(excludeIDsStr, 10, 32); err == nil {
			excludeIDs = append(excludeIDs, uint(id))
		}
	}
	meals, err := services.AllMeals(h.db, excludeIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"meals": meals})
}

func (h *MealHandler) getMeal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	meal, err := services.MealByID(h.db, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meal)
}

type CreateMealRequest struct {
	Meal models.Meal `json:"meal"`
	Log  bool        `json:"log"`
}

func (h *MealHandler) postNewMeal(c *gin.Context) {
	var req CreateMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	mealID, err := services.CreateMeal(h.db, &req.Meal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Log {
		day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if day == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Day not found"})
			return
		}
		if err := services.CreateDayMeal(h.db, &models.DayLog{
			DayID:  day.ID,
			MealID: mealID,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"meal_id": mealID})
}

type LogPlannedMealRequest struct {
	MealID uint `json:"meal_id"`
}

func (h *MealHandler) postLogPlanned(c *gin.Context) {
	var req LogPlannedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if day == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Day not found"})
		return
	}
	if err := services.SetPlannedMealLogged(h.db, day.ID, req.MealID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": totalCalories,
		"totalProtein":  totalProtein,
		"totalFiber":    totalFiber,
		"totalCarbs":    totalCarbs,
	})
}

type EditLoggedMealRequest struct {
	Meal      models.Meal `json:"meal"`
	OldMealID uint        `json:"oldMealID"`
}

func (h *MealHandler) postLogEdited(c *gin.Context) {
	var req EditLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newMealID, err := services.CreateMeal(h.db, &req.Meal)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if day == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Day not found"})
		return
	}

	if err := services.CreateDayMeal(h.db, &models.DayLog{
		DayID:  day.ID,
		MealID: newMealID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": totalCalories,
		"totalProtein":  totalProtein,
		"totalFiber":    totalFiber,
		"totalCarbs":    totalCarbs,
	})
}

func (h *MealHandler) postEditLogged(c *gin.Context) {
	var req EditLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if day == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Day not found"})
		return
	}
	if err := services.UpdateDayLogMeal(h.db, day.ID, req.OldMealID, req.Meal.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": totalCalories,
		"totalProtein":  totalProtein,
		"totalFiber":    totalFiber,
		"totalCarbs":    totalCarbs,
	})
}

type DeleteLoggedMealRequest struct {
	MealID uint `json:"meal_id"`
}

func (h *MealHandler) deleteLoggedMeal(c *gin.Context) {
	var req DeleteLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if day == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Day not found"})
		return
	}
	if err := services.DeleteLoggedMeal(h.db, day.ID, req.MealID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": totalCalories,
		"totalProtein":  totalProtein,
		"totalFiber":    totalFiber,
		"totalCarbs":    totalCarbs,
	})
}
