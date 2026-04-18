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
		meals.GET("/saved-meal/all", h.getAllSavedMeals)
		meals.POST("/saved-meal/new", h.postNewSavedMeal)
		meals.GET("/meal/:id", h.getMeal)
		meals.POST("/meal/new", h.postNewMeal)
		meals.POST("/meal/log-planned", h.postLogPlanned)
		meals.POST("/meal/logedited", h.postLogEdited)
		meals.POST("/meal/editlogged", h.postEditLogged)
		meals.DELETE("/meal/logged", h.deleteLoggedMeal)
		meals.POST("/planned/from-saved", h.postPlannedFromSaved)
		meals.DELETE("/planned", h.deletePlannedMeal)
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

func (h *MealHandler) getAllSavedMeals(c *gin.Context) {
	excludeIDsStr := c.Query("exclude")
	var excludeIDs []uint
	if excludeIDsStr != "" {
		if id, err := strconv.ParseUint(excludeIDsStr, 10, 32); err == nil {
			excludeIDs = append(excludeIDs, uint(id))
		}
	}
	saved, err := services.AllSavedMeals(h.db, excludeIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"saved_meals": saved})
}

func (h *MealHandler) postNewSavedMeal(c *gin.Context) {
	var sm models.SavedMeal
	if err := c.ShouldBindJSON(&sm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sm.ID = 0
	id, err := services.CreateSavedMeal(h.db, &sm)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"saved_meal_id": id})
}

func savedMealFromMealTemplate(m *models.Meal) *models.SavedMeal {
	s := &models.SavedMeal{Name: m.Name}
	for _, it := range m.Items {
		s.Items = append(s.Items, models.SavedMealItem{
			FoodID: it.FoodID,
			Amount: float64(it.Amount),
		})
	}
	return s
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
	Meal          models.Meal `json:"meal"`
	Log           bool        `json:"log"`
	SaveToLibrary bool        `json:"save_to_library"`
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

	if req.Log && req.SaveToLibrary {
		sm := savedMealFromMealTemplate(&req.Meal)
		if _, err := services.CreateSavedMeal(h.db, sm); err != nil {
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
	// Client sends meal with ID 0 (same as postNewMeal): persist edited meal, then point day_log at the new row.
	var newMealID uint
	err = h.db.Transaction(func(tx *gorm.DB) error {
		req.Meal.ID = 0
		var createErr error
		newMealID, createErr = services.CreateMeal(tx, &req.Meal)
		if createErr != nil {
			return createErr
		}
		return services.UpdateDayLogMeal(tx, day.ID, req.OldMealID, newMealID)
	})
	if err != nil {
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

type AddPlannedFromSavedRequest struct {
	SavedMealID uint `json:"saved_meal_id"`
	Offset      int  `json:"offset"`
}

func (h *MealHandler) postPlannedFromSaved(c *gin.Context) {
	var req AddPlannedFromSavedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.SavedMealID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "saved_meal_id is required"})
		return
	}
	if err := services.AddPlannedMealFromSavedMeal(h.db, req.Offset, req.SavedMealID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(req.Offset))
	if err != nil || day == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Day not found"})
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

type DeletePlannedMealRequest struct {
	PlannedMealID uint `json:"planned_meal_id"`
	Offset        int  `json:"offset"`
}

func (h *MealHandler) deletePlannedMeal(c *gin.Context) {
	var req DeletePlannedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.PlannedMealID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "planned_meal_id is required"})
		return
	}
	if err := services.DeletePlannedMeal(h.db, req.Offset, req.PlannedMealID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(req.Offset))
	if err != nil || day == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Day not found"})
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
