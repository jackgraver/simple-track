package controller

import (
	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
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

func dayWithTotalsResponse(result services.DayWithTotals) gin.H {
	return gin.H{
		"day":           result.Day,
		"totalCalories": result.Totals.Calories,
		"totalProtein":  result.Totals.Protein,
		"totalFiber":    result.Totals.Fiber,
		"totalCarbs":    result.Totals.Carbs,
		"totalFat":      result.Totals.Fat,
	}
}

func PostQuickLog(c *gin.Context) {
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
	result, err := services.QuickLogMeal(dietrepo.QuickLogParams{
		DisplayName:   displayName,
		FoodRowName:   fmt.Sprintf("%s [ql-%d]", displayName, time.Now().UnixNano()),
		Calories:      req.Calories,
		Protein:       req.Protein,
		Carbs:         req.Carbs,
		Fat:           req.Fat,
		Fiber:         req.Fiber,
		Offset:        req.Offset,
		ReplaceMealID: req.ReplaceMealID,
	})
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

func PostFood(c *gin.Context) {
	var req CreateFoodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	createdFood, err := services.CreateFood(&req.Food, req.RelatedFoodID)
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

func parseExcludeIDs(c *gin.Context) []uint {
	excludeIDsStr := c.Query("exclude")
	var excludeIDs []uint
	if excludeIDsStr != "" {
		if id, err := strconv.ParseUint(excludeIDsStr, 10, 32); err == nil {
			excludeIDs = append(excludeIDs, uint(id))
		}
	}
	return excludeIDs
}

func GetAllFoods(c *gin.Context) {
	excludeIDs := parseExcludeIDs(c)
	foods, err := services.AllFoodsForPicker(excludeIDs)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	composites, err := services.AllCompositeFoods()
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

func PostNewCompositeFood(c *gin.Context) {
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
	id, err := services.CreateCompositeFood(&cf)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	loaded, err := services.CompositeFoodByID(id)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"composite_food": compositeToResponse(*loaded)})
}

func GetAllMeals(c *gin.Context) {
	meals, err := services.AllMeals(parseExcludeIDs(c))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"meals": meals})
}

func GetAllSavedMeals(c *gin.Context) {
	saved, err := services.AllSavedMeals(parseExcludeIDs(c))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"saved_meals": saved})
}

func PostNewSavedMeal(c *gin.Context) {
	var sm models.SavedMeal
	if err := c.ShouldBindJSON(&sm); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	sm.ID = 0
	id, err := services.CreateSavedMeal(&sm)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"saved_meal_id": id})
}

func DeleteSavedMeal(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		apierr.BadRequest(c, "invalid saved meal id")
		return
	}
	id := uint(id64)
	force := strings.ToLower(strings.TrimSpace(c.Query("force"))) == "true"
	if _, err := services.SavedMealByID(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		apierr.Internal(c, err)
		return
	}
	if !force {
		info, err := services.SavedMealPlannedUsageInfo(id)
		if err != nil {
			apierr.Internal(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"reference_count": info.ReferenceCount,
		})
		return
	}
	if err := services.DeleteSavedMeal(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		apierr.Internal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func GetSavedMeal(c *gin.Context) {
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id64 == 0 {
		apierr.BadRequest(c, "invalid saved meal id")
		return
	}
	sm, err := services.SavedMealByID(uint(id64))
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

func PutSavedMeal(c *gin.Context) {
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
	if err := services.ReplaceSavedMeal(uint(id64), &body); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		apierr.Internal(c, err)
		return
	}
	updated, err := services.SavedMealByID(uint(id64))
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

func GetMeal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		apierr.BadRequest(c, "Invalid ID")
		return
	}
	meal, err := services.MealByID(uint(id))
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
	Offset        int         `json:"offset"`
}

func PostNewMeal(c *gin.Context) {
	var req CreateMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	mealID, err := services.CreateMeal(&req.Meal)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if req.Log {
		day, err := services.FindMealPlanDay(utils.ZerodTime(req.Offset))
		if err != nil {
			apierr.Internal(c, err)
			return
		}
		if day == nil {
			apierr.NotFound(c, "Day not found")
			return
		}
		if err := services.CreateDayMeal(&models.DayLog{
			DayID:  day.ID,
			MealID: mealID,
		}); err != nil {
			apierr.Internal(c, err)
			return
		}
	}
	if req.Log && req.SaveToLibrary {
		sm := savedMealFromMealTemplate(&req.Meal)
		if _, err := services.CreateSavedMeal(sm); err != nil {
			apierr.Internal(c, err)
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"meal_id": mealID})
}

type LogPlannedMealRequest struct {
	MealID uint `json:"meal_id"`
}

func PostLogPlanned(c *gin.Context) {
	var req LogPlannedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	day, err := services.FindMealPlanDay(utils.ZerodTime(0))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	if err := services.SetPlannedMealLogged(day.ID, req.MealID); err != nil {
		apierr.Internal(c, err)
		return
	}
	exists, err := services.DayLogExistsForMeal(day.ID, req.MealID)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if !exists {
		if err := services.CreateDayMeal(&models.DayLog{
			DayID:  day.ID,
			MealID: req.MealID,
		}); err != nil {
			apierr.Internal(c, err)
			return
		}
	}
	result, err := services.ReloadDayWithTotals(day)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

type EditLoggedMealRequest struct {
	Meal                models.Meal `json:"meal"`
	OldMealID           uint        `json:"oldMealID"`
	PlannedSourceMealID uint        `json:"planned_source_meal_id"`
	DayID               uint        `json:"day_id"`
}

func PostLogEdited(c *gin.Context) {
	var req EditLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	newMealID, err := services.CreateMeal(&req.Meal)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	day, err := services.FindMealPlanDayByRowIDOrToday(req.DayID)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	if err := services.CreateDayMeal(&models.DayLog{
		DayID:  day.ID,
		MealID: newMealID,
	}); err != nil {
		apierr.Internal(c, err)
		return
	}
	if req.PlannedSourceMealID != 0 {
		if err := services.SetPlannedMealLogged(day.ID, req.PlannedSourceMealID); err != nil {
			apierr.Internal(c, err)
			return
		}
	}
	result, err := services.ReloadDayWithTotals(day)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

func PostEditLogged(c *gin.Context) {
	var req EditLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	day, err := services.FindMealPlanDayByRowIDOrToday(req.DayID)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	if err := services.EditLoggedMeal(day.ID, req.OldMealID, &req.Meal); err != nil {
		apierr.Internal(c, err)
		return
	}
	result, err := services.ReloadDayWithTotals(day)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

type DeleteLoggedMealRequest struct {
	MealID uint `json:"meal_id"`
	DayID  uint `json:"day_id"`
}

func DeleteLoggedMeal(c *gin.Context) {
	var req DeleteLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	day, err := services.FindMealPlanDayByRowIDOrToday(req.DayID)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	if err := services.DeleteLoggedMeal(day.ID, req.MealID); err != nil {
		apierr.Internal(c, err)
		return
	}
	result, err := services.ReloadDayWithTotals(day)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

type AddPlannedFromSavedRequest struct {
	SavedMealID uint `json:"saved_meal_id"`
	Offset      int  `json:"offset"`
}

func PostPlannedFromSaved(c *gin.Context) {
	var req AddPlannedFromSavedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if req.SavedMealID == 0 {
		apierr.BadRequest(c, "saved_meal_id is required")
		return
	}
	if err := services.AddPlannedMealFromSavedMeal(req.Offset, req.SavedMealID); err != nil {
		apierr.Internal(c, err)
		return
	}
	day, err := services.FindMealPlanDay(utils.ZerodTime(req.Offset))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	result, err := services.ReloadDayWithTotals(day)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

type AddPlannedFromLabelRequest struct {
	Name   string `json:"name"`
	Offset int    `json:"offset"`
}

func PostPlannedFromLabel(c *gin.Context) {
	var req AddPlannedFromLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		apierr.BadRequest(c, "name is required")
		return
	}
	if err := services.AddPlannedMealFromLabel(req.Offset, name); err != nil {
		apierr.Internal(c, err)
		return
	}
	day, err := services.FindMealPlanDay(utils.ZerodTime(req.Offset))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	result, err := services.ReloadDayWithTotals(day)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

type ReorderPlannedMealsRequest struct {
	Offset         int    `json:"offset"`
	PlannedMealIDs []uint `json:"planned_meal_ids"`
}

func PostPlannedReorder(c *gin.Context) {
	var req ReorderPlannedMealsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if len(req.PlannedMealIDs) == 0 {
		apierr.BadRequest(c, "planned_meal_ids is required")
		return
	}
	if err := services.ReorderPlannedMeals(req.Offset, req.PlannedMealIDs); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	day, err := services.FindMealPlanDay(utils.ZerodTime(req.Offset))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	result, err := services.ReloadDayWithTotals(day)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

type DeletePlannedMealRequest struct {
	PlannedMealID uint `json:"planned_meal_id"`
	Offset        int  `json:"offset"`
}

func DeletePlannedMeal(c *gin.Context) {
	var req DeletePlannedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if req.PlannedMealID == 0 {
		apierr.BadRequest(c, "planned_meal_id is required")
		return
	}
	if err := services.DeletePlannedMeal(req.Offset, req.PlannedMealID); err != nil {
		apierr.Internal(c, err)
		return
	}
	day, err := services.FindMealPlanDay(utils.ZerodTime(req.Offset))
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	result, err := services.ReloadDayWithTotals(day)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}
