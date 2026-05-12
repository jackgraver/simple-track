package routes

import (
	"be-simpletracker/internal/core/diet/models"
	"be-simpletracker/internal/core/diet/services"
	"be-simpletracker/internal/utils"
	"be-simpletracker/internal/utils/apierr"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		meals.POST("/composite-food/new", h.postNewCompositeFood)
		meals.GET("/meal/all", h.getAllMeals)
		meals.GET("/saved-meal/all", h.getAllSavedMeals)
		meals.GET("/saved-meal/:id", h.getSavedMeal)
		meals.POST("/saved-meal/new", h.postNewSavedMeal)
		meals.PUT("/saved-meal/:id", h.putSavedMeal)
		meals.DELETE("/saved-meal/:id", h.deleteSavedMeal)
		meals.GET("/meal/:id", h.getMeal)
		meals.POST("/quick-log", h.postQuickLog)
		meals.POST("/meal/new", h.postNewMeal)
		meals.POST("/meal/log-planned", h.postLogPlanned)
		meals.POST("/meal/logedited", h.postLogEdited)
		meals.POST("/meal/editlogged", h.postEditLogged)
		meals.DELETE("/meal/logged", h.deleteLoggedMeal)
		meals.POST("/planned/from-saved", h.postPlannedFromSaved)
		meals.POST("/planned/reorder", h.postPlannedReorder)
		meals.DELETE("/planned", h.deletePlannedMeal)
	}
}

type CreateFoodRequest struct {
	models.Food
	RelatedFoodID *uint `json:"related_food_id,omitempty"`
}

type QuickLogRequest struct {
	Name          string  `json:"name" binding:"required"`
	Calories      float32 `json:"calories"`
	Protein       float32 `json:"protein"`
	Carbs         float32 `json:"carbs"`
	Fat           float32 `json:"fat"`
	Fiber         float32 `json:"fiber"`
	Offset        int     `json:"offset"`
	ReplaceMealID uint    `json:"replace_meal_id"`
}

func (h *MealHandler) postQuickLog(c *gin.Context) {
	var req QuickLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	displayName := strings.TrimSpace(req.Name)
	if displayName == "" {
		apierr.BadRequest(c, "name is required")
		return
	}
	foodRowName := fmt.Sprintf("%s [ql-%d]", displayName, time.Now().UnixNano())
	var dayID uint
	err := h.db.Transaction(func(tx *gorm.DB) error {
		day, derr := services.FindMealPlanDay(tx, utils.ZerodTime(req.Offset))
		if derr != nil {
			return derr
		}
		if day == nil {
			return fmt.Errorf("day not found")
		}
		dayID = day.ID
		food := models.Food{
			Name:          foodRowName,
			ServingType:   "",
			ServingAmount: 1,
			Calories:      req.Calories,
			Protein:       req.Protein,
			Fiber:         req.Fiber,
			Carbs:         req.Carbs,
			Fat:           req.Fat,
			QuickEntry:    true,
		}
		if err := tx.Create(&food).Error; err != nil {
			return err
		}
		meal := models.Meal{
			Name: displayName,
			Items: []models.MealItem{{
				FoodID: food.ID,
				Amount: 1,
			}},
		}
		mealID, merr := services.CreateMeal(tx, &meal)
		if merr != nil {
			return merr
		}
		if req.ReplaceMealID != 0 {
			var cnt int64
			if cerr := tx.Model(&models.DayLog{}).
				Where("day_id = ? AND meal_id = ? AND deleted_at IS NULL", dayID, req.ReplaceMealID).
				Count(&cnt).Error; cerr != nil {
				return cerr
			}
			if cnt != 1 {
				return fmt.Errorf("replace_meal_id does not match a log on this day")
			}
			return services.UpdateDayLogMeal(tx, dayID, req.ReplaceMealID, mealID)
		}
		return services.CreateDayMeal(tx, &models.DayLog{
			DayID:  dayID,
			MealID: mealID,
		})
	})
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	day, err := services.MealPlanDayByID(h.db, int(dayID))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	tot := services.CalculateTotals(h.db, dayID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
	})
}

func (h *MealHandler) postFood(c *gin.Context) {
	var req CreateFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	createdFood, err := services.CreateFood(h.db, &req.Food, req.RelatedFoodID)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"food": createdFood})
}

type compositeFoodWithMacros struct {
	ID        uint                       `json:"ID"`
	CreatedAt time.Time                  `json:"created_at"`
	UpdatedAt time.Time                  `json:"updated_at"`
	Name      string                     `json:"name"`
	Items     []models.CompositeFoodItem `json:"items"`
	EntryKind string                     `json:"entry_kind"`
	Calories  float32                    `json:"calories"`
	Protein   float32                    `json:"protein"`
	Fiber     float32                    `json:"fiber"`
	Carbs     float32                    `json:"carbs"`
	Fat       float32                    `json:"fat"`
}

func sumCompositeMacros(cf models.CompositeFood) (cal, pro, fib, carb, fat float32) {
	for _, it := range cf.Items {
		a := it.Amount
		cal += it.Food.Calories * a
		pro += it.Food.Protein * a
		fib += it.Food.Fiber * a
		carb += it.Food.Carbs * a
		fat += it.Food.Fat * a
	}
	return cal, pro, fib, carb, fat
}

func compositeToResponse(cf models.CompositeFood) compositeFoodWithMacros {
	cal, pro, fib, carb, fat := sumCompositeMacros(cf)
	return compositeFoodWithMacros{
		ID:        cf.ID,
		CreatedAt: cf.CreatedAt,
		UpdatedAt: cf.UpdatedAt,
		Name:      cf.Name,
		Items:     cf.Items,
		EntryKind: "composite",
		Calories:  cal,
		Protein:   pro,
		Fiber:     fib,
		Carbs:     carb,
		Fat:       fat,
	}
}

func (h *MealHandler) getAllFoods(c *gin.Context) {
	excludeIDsStr := c.Query("exclude")
	var excludeIDs []uint
	if excludeIDsStr != "" {
		if id, err := strconv.ParseUint(excludeIDsStr, 10, 32); err == nil {
			excludeIDs = append(excludeIDs, uint(id))
		}
	}
	foods, err := services.AllFoodsForPicker(h.db, excludeIDs)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	composites, err := services.AllCompositeFoods(h.db)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	compositeDTOs := make([]compositeFoodWithMacros, 0, len(composites))
	for _, cf := range composites {
		compositeDTOs = append(compositeDTOs, compositeToResponse(cf))
	}
	c.JSON(http.StatusOK, gin.H{
		"foods":           foods,
		"composite_foods": compositeDTOs,
	})
}

func (h *MealHandler) postNewCompositeFood(c *gin.Context) {
	var cf models.CompositeFood
	if err := c.ShouldBindJSON(&cf); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	cf.ID = 0
	if cf.Name == "" || len(cf.Items) == 0 {
		apierr.BadRequest(c, "name and at least one item required")
		return
	}
	id, err := services.CreateCompositeFood(h.db, &cf)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	var loaded models.CompositeFood
	if err := h.db.Preload("Items.Food").First(&loaded, id).Error; err != nil {
		apierr.Internal(c, err)
		return
	}
	for i := range loaded.Items {
		models.NormalizeQuickLogFoodNameForResponse(&loaded.Items[i].Food)
	}
	c.JSON(http.StatusCreated, gin.H{"composite_food": compositeToResponse(loaded)})
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
		apierr.Internal(c, err)
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
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"saved_meals": saved})
}

func (h *MealHandler) postNewSavedMeal(c *gin.Context) {
	var sm models.SavedMeal
	if err := c.ShouldBindJSON(&sm); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	sm.ID = 0
	id, err := services.CreateSavedMeal(h.db, &sm)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"saved_meal_id": id})
}

func (h *MealHandler) deleteSavedMeal(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		apierr.BadRequest(c, "invalid saved meal id")
		return
	}
	id := uint(id64)
	force := strings.ToLower(strings.TrimSpace(c.Query("force"))) == "true"

	if _, err := services.SavedMealByID(h.db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		apierr.Internal(c, err)
		return
	}

	if !force {
		info, err := services.SavedMealPlannedUsageInfo(h.db, id)
		if err != nil {
			apierr.Internal(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"reference_count": info.ReferenceCount,
		})
		return
	}

	if err := services.DeleteSavedMeal(h.db, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		apierr.Internal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MealHandler) getSavedMeal(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		apierr.BadRequest(c, "invalid saved meal id")
		return
	}
	sm, err := services.SavedMealByID(h.db, uint(id64))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		apierr.Internal(c, err)
		return
	}
	for i := range sm.Items {
		models.NormalizeQuickLogFoodNameForResponse(&sm.Items[i].Food)
	}
	c.JSON(http.StatusOK, gin.H{"saved_meal": sm})
}

func (h *MealHandler) putSavedMeal(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		apierr.BadRequest(c, "invalid saved meal id")
		return
	}
	var body models.SavedMeal
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	body.ID = 0
	if body.Name == "" || len(body.Items) == 0 {
		apierr.BadRequest(c, "name and at least one item required")
		return
	}
	if err := services.ReplaceSavedMeal(h.db, uint(id64), &body); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		apierr.Internal(c, err)
		return
	}
	updated, err := services.SavedMealByID(h.db, uint(id64))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	for i := range updated.Items {
		models.NormalizeQuickLogFoodNameForResponse(&updated.Items[i].Food)
	}
	c.JSON(http.StatusOK, gin.H{"saved_meal": updated})
}

func savedMealFromMealTemplate(m *models.Meal) *models.SavedMeal {
	s := &models.SavedMeal{Name: m.Name}
	for _, it := range m.Items {
		s.Items = append(s.Items, models.SavedMealItem{
			FoodID:          it.FoodID,
			Amount:          float64(it.Amount),
			GroupID:         it.GroupID,
			GroupLabel:      it.GroupLabel,
			CompositeFoodID: it.CompositeFoodID,
		})
	}
	return s
}

func (h *MealHandler) getMeal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid ID")
		return
	}
	meal, err := services.MealByID(h.db, uint(id))
	if err != nil {
		apierr.Internal(c, err)
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
		apierr.BadRequest(c, err.Error())
		return
	}
	mealID, err := services.CreateMeal(h.db, &req.Meal)
	if err != nil {
		apierr.Internal(c, err)
		return
	}

	if req.Log {
		day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
		if err != nil {
			apierr.Internal(c, err)
			return
		}
		if day == nil {
			apierr.NotFound(c, "Day not found")
			return
		}
		if err := services.CreateDayMeal(h.db, &models.DayLog{
			DayID:  day.ID,
			MealID: mealID,
		}); err != nil {
			apierr.Internal(c, err)
			return
		}
	}

	if req.Log && req.SaveToLibrary {
		sm := savedMealFromMealTemplate(&req.Meal)
		if _, err := services.CreateSavedMeal(h.db, sm); err != nil {
			apierr.Internal(c, err)
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
		apierr.BadRequest(c, err.Error())
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	if err := services.SetPlannedMealLogged(h.db, day.ID, req.MealID); err != nil {
		apierr.Internal(c, err)
		return
	}
	exists, err := services.DayLogExistsForMeal(h.db, day.ID, req.MealID)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if !exists {
		if err := services.CreateDayMeal(h.db, &models.DayLog{
			DayID:  day.ID,
			MealID: req.MealID,
		}); err != nil {
			apierr.Internal(c, err)
			return
		}
	}
	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	tot := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
	})
}

type EditLoggedMealRequest struct {
	Meal                models.Meal `json:"meal"`
	OldMealID           uint        `json:"oldMealID"`
	PlannedSourceMealID uint        `json:"planned_source_meal_id"`
}

func (h *MealHandler) postLogEdited(c *gin.Context) {
	var req EditLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	newMealID, err := services.CreateMeal(h.db, &req.Meal)
	if err != nil {
		apierr.Internal(c, err)
		return
	}

	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}

	if err := services.CreateDayMeal(h.db, &models.DayLog{
		DayID:  day.ID,
		MealID: newMealID,
	}); err != nil {
		apierr.Internal(c, err)
		return
	}

	if req.PlannedSourceMealID != 0 {
		if err := services.SetPlannedMealLogged(h.db, day.ID, req.PlannedSourceMealID); err != nil {
			apierr.Internal(c, err)
			return
		}
	}

	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	tot := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
	})
}

func (h *MealHandler) postEditLogged(c *gin.Context) {
	var req EditLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
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
		apierr.Internal(c, err)
		return
	}
	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	tot := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
	})
}

type DeleteLoggedMealRequest struct {
	MealID uint `json:"meal_id"`
}

func (h *MealHandler) deleteLoggedMeal(c *gin.Context) {
	var req DeleteLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	if err := services.DeleteLoggedMeal(h.db, day.ID, req.MealID); err != nil {
		apierr.Internal(c, err)
		return
	}
	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	tot := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
	})
}

type AddPlannedFromSavedRequest struct {
	SavedMealID uint `json:"saved_meal_id"`
	Offset      int  `json:"offset"`
}

func (h *MealHandler) postPlannedFromSaved(c *gin.Context) {
	var req AddPlannedFromSavedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if req.SavedMealID == 0 {
		apierr.BadRequest(c, "saved_meal_id is required")
		return
	}
	if err := services.AddPlannedMealFromSavedMeal(h.db, req.Offset, req.SavedMealID); err != nil {
		apierr.Internal(c, err)
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(req.Offset))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	tot := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
	})
}

type ReorderPlannedMealsRequest struct {
	Offset         int    `json:"offset"`
	PlannedMealIDs []uint `json:"planned_meal_ids"`
}

func (h *MealHandler) postPlannedReorder(c *gin.Context) {
	var req ReorderPlannedMealsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if len(req.PlannedMealIDs) == 0 {
		apierr.BadRequest(c, "planned_meal_ids is required")
		return
	}
	if err := services.ReorderPlannedMeals(h.db, req.Offset, req.PlannedMealIDs); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(req.Offset))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	tot := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
	})
}

type DeletePlannedMealRequest struct {
	PlannedMealID uint `json:"planned_meal_id"`
	Offset        int  `json:"offset"`
}

func (h *MealHandler) deletePlannedMeal(c *gin.Context) {
	var req DeletePlannedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if req.PlannedMealID == 0 {
		apierr.BadRequest(c, "planned_meal_id is required")
		return
	}
	if err := services.DeletePlannedMeal(h.db, req.Offset, req.PlannedMealID); err != nil {
		apierr.Internal(c, err)
		return
	}
	day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(req.Offset))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	day, err = services.MealPlanDayByID(h.db, int(day.ID))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	tot := services.CalculateTotals(h.db, day.ID)
	c.JSON(http.StatusOK, gin.H{
		"day":           day,
		"totalCalories": tot.Calories,
		"totalProtein":  tot.Protein,
		"totalFiber":    tot.Fiber,
		"totalCarbs":    tot.Carbs,
		"totalFat":      tot.Fat,
	})
}
