package controller

import (
	"be-simpletracker/internal/clock"
	"be-simpletracker/internal/core/diet/models"
	dietrepo "be-simpletracker/internal/core/diet/repository"
	"be-simpletracker/internal/core/diet/services"
	"be-simpletracker/internal/utils/apierr"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// normalizeClientID accepts an empty client_id (the request is not idempotent) or a uuid, which
// it returns in canonical lowercase form. Anything else is rejected with a 400 instead of
// reaching the uuid column and failing there with a 500.
func normalizeClientID(c *gin.Context, s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", true
	}
	id, err := uuid.Parse(s)
	if err != nil || id == uuid.Nil {
		apierr.BadRequest(c, "client_id must be a uuid")
		return "", false
	}
	return id.String(), true
}

// Request bodies are their own types, never gorm models (BACKLOG #1): a body can't set user_id,
// created_by, timestamps or a nested food, and ingredient lines point at foods by ID only.

type foodBody struct {
	Name          string  `json:"name"`
	ServingType   string  `json:"serving_type"`
	ServingAmount float32 `json:"serving_amount"`
	Calories      float32 `json:"calories"`
	Protein       float32 `json:"protein"`
	Fiber         float32 `json:"fiber"`
	Carbs         float32 `json:"carbs"`
	Fat           float32 `json:"fat"`
}

func (b foodBody) validate() string {
	switch {
	case strings.TrimSpace(b.Name) == "":
		return "name is required"
	case b.ServingAmount < 0 || b.Calories < 0 || b.Protein < 0 || b.Fiber < 0 || b.Carbs < 0 || b.Fat < 0:
		return "serving and macro values must not be negative"
	}
	return ""
}

type CreateFoodRequest struct {
	foodBody
	RelatedFoodID *uint `json:"related_food_id,omitempty"`
}

type mealItemBody struct {
	FoodID          uint    `json:"food_id"`
	Amount          float32 `json:"amount"`
	GroupID         string  `json:"group_id"`
	GroupLabel      string  `json:"group_label"`
	CompositeFoodID *uint   `json:"composite_food_id"`
}

type mealBody struct {
	Name     string         `json:"name"`
	Items    []mealItemBody `json:"items"`
	ClientID string         `json:"client_id"`
}

// toModel builds the meal to create. The client_id has been through normalizeClientID by the
// caller (empty means not idempotent).
func (b mealBody) toModel(clientID string) models.Meal {
	m := models.Meal{Name: b.Name, ClientID: nilIfEmpty(clientID)}
	for _, it := range b.Items {
		m.Items = append(m.Items, models.MealItem{
			FoodID: it.FoodID, Amount: it.Amount, GroupID: it.GroupID,
			GroupLabel: it.GroupLabel, CompositeFoodID: it.CompositeFoodID,
		})
	}
	return m
}

type savedMealItemBody struct {
	FoodID          uint    `json:"food_id"`
	Amount          float64 `json:"amount"`
	GroupID         string  `json:"group_id"`
	GroupLabel      string  `json:"group_label"`
	CompositeFoodID *uint   `json:"composite_food_id"`
}

type savedMealBody struct {
	Name  string              `json:"name"`
	Items []savedMealItemBody `json:"items"`
}

func (b savedMealBody) toModel() models.SavedMeal {
	sm := models.SavedMeal{Name: b.Name}
	for _, it := range b.Items {
		sm.Items = append(sm.Items, models.SavedMealItem{
			FoodID: it.FoodID, Amount: it.Amount, GroupID: it.GroupID,
			GroupLabel: it.GroupLabel, CompositeFoodID: it.CompositeFoodID,
		})
	}
	return sm
}

type compositeFoodBody struct {
	Name  string `json:"name"`
	Items []struct {
		FoodID uint    `json:"food_id"`
		Amount float32 `json:"amount"`
	} `json:"items"`
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// respondErr answers an error from the diet services: our own sentinels with their messages, a
// missing or someone else's row as the shared 404, anything else as a logged 500.
func respondErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, dietrepo.ErrUnknownFood):
		apierr.BadRequest(c, "unknown food or recipe in the request")
	case errors.Is(err, dietrepo.ErrForbidden):
		apierr.Forbidden(c, "only the person who added this, or an admin, can do that")
	case errors.Is(err, dietrepo.ErrFoodNameTaken), errors.Is(err, dietrepo.ErrCompositeNameTaken):
		apierr.Conflict(c, err.Error())
	case errors.Is(err, dietrepo.ErrInvalidRequest):
		apierr.BadRequest(c, err.Error())
	default:
		apierr.NotFoundOrInternal(c, err)
	}
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
	ClientID      string  `json:"client_id"`
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
	clientID, ok := normalizeClientID(c, req.ClientID)
	if !ok {
		return
	}
	result, err := services.QuickLogMeal(c.Request.Context(), dietrepo.QuickLogParams{
		DisplayName:   displayName,
		FoodRowName:   fmt.Sprintf("%s [ql-%d]", displayName, time.Now().UnixNano()),
		Calories:      req.Calories,
		Protein:       req.Protein,
		Carbs:         req.Carbs,
		Fat:           req.Fat,
		Fiber:         req.Fiber,
		Offset:        req.Offset,
		ReplaceMealID: req.ReplaceMealID,
		ClientID:      clientID,
	})
	if err != nil {
		respondErr(c, err)
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
	if msg := req.validate(); msg != "" {
		apierr.BadRequest(c, msg)
		return
	}
	food := models.Food{
		Name: strings.TrimSpace(req.Name), ServingType: req.ServingType, ServingAmount: req.ServingAmount,
		Calories: req.Calories, Protein: req.Protein, Fiber: req.Fiber, Carbs: req.Carbs, Fat: req.Fat,
	}
	createdFood, err := services.CreateFood(c.Request.Context(), &food, req.RelatedFoodID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"food": createdFood})
}

func parseFoodID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		apierr.BadRequest(c, "invalid food id")
		return 0, false
	}
	return uint(id), true
}

// PutFood edits a catalog food: its creator or an admin (403 for anyone else).
func PutFood(c *gin.Context) {
	id, ok := parseFoodID(c)
	if !ok {
		return
	}
	var req foodBody
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid body")
		return
	}
	if msg := req.validate(); msg != "" {
		apierr.BadRequest(c, msg)
		return
	}
	food, err := services.UpdateFood(c.Request.Context(), id, dietrepo.FoodFields{
		Name: strings.TrimSpace(req.Name), ServingType: req.ServingType, ServingAmount: req.ServingAmount,
		Calories: req.Calories, Protein: req.Protein, Fiber: req.Fiber, Carbs: req.Carbs, Fat: req.Fat,
	})
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"food": food})
}

// DeleteFood removes a food from the shared catalog: admin only (403 for anyone else). Meals that
// used it keep showing it.
func DeleteFood(c *gin.Context) {
	id, ok := parseFoodID(c)
	if !ok {
		return
	}
	if err := services.DeleteFood(c.Request.Context(), id); err != nil {
		respondErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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
	foods, err := services.AllFoodsForPicker(c.Request.Context(), excludeIDs)
	if err != nil {
		respondErr(c, err)
		return
	}
	composites, err := services.AllCompositeFoods(c.Request.Context())
	if err != nil {
		respondErr(c, err)
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
	var body compositeFoodBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "invalid body")
		return
	}
	if strings.TrimSpace(body.Name) == "" || len(body.Items) == 0 {
		apierr.BadRequest(c, "name and at least one item required")
		return
	}
	cf := models.CompositeFood{Name: strings.TrimSpace(body.Name)}
	for _, it := range body.Items {
		cf.Items = append(cf.Items, models.CompositeFoodItem{FoodID: it.FoodID, Amount: it.Amount})
	}
	id, err := services.CreateCompositeFood(c.Request.Context(), &cf)
	if err != nil {
		respondErr(c, err)
		return
	}
	loaded, err := services.CompositeFoodByID(c.Request.Context(), id)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"composite_food": compositeToResponse(*loaded)})
}

func GetAllMeals(c *gin.Context) {
	meals, err := services.AllMeals(c.Request.Context(), parseExcludeIDs(c))
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"meals": meals})
}

func GetAllSavedMeals(c *gin.Context) {
	saved, err := services.AllSavedMeals(c.Request.Context(), parseExcludeIDs(c))
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"saved_meals": saved})
}

func PostNewSavedMeal(c *gin.Context) {
	var body savedMealBody
	if err := c.ShouldBindJSON(&body); err != nil {
		apierr.BadRequest(c, "invalid body")
		return
	}
	sm := body.toModel()
	id, err := services.CreateSavedMeal(c.Request.Context(), &sm)
	if err != nil {
		respondErr(c, err)
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
	if _, err := services.SavedMealByID(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		respondErr(c, err)
		return
	}
	if !force {
		info, err := services.SavedMealPlannedUsageInfo(c.Request.Context(), id)
		if err != nil {
			respondErr(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"reference_count": info.ReferenceCount,
		})
		return
	}
	if err := services.DeleteSavedMeal(c.Request.Context(), id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		respondErr(c, err)
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
	sm, err := services.SavedMealByID(c.Request.Context(), uint(id64))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		respondErr(c, err)
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
	var req savedMealBody
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "invalid body")
		return
	}
	if req.Name == "" || len(req.Items) == 0 {
		apierr.BadRequest(c, "name and at least one item required")
		return
	}
	body := req.toModel()
	if err := services.ReplaceSavedMeal(c.Request.Context(), uint(id64), &body); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "saved meal not found")
			return
		}
		respondErr(c, err)
		return
	}
	updated, err := services.SavedMealByID(c.Request.Context(), uint(id64))
	if err != nil {
		respondErr(c, err)
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
	meal, err := services.MealByID(c.Request.Context(), uint(id))
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, meal)
}

type CreateMealRequest struct {
	Meal          mealBody `json:"meal"`
	Log           bool     `json:"log"`
	SaveToLibrary bool     `json:"save_to_library"`
	Offset        int      `json:"offset"`
}

func PostNewMeal(c *gin.Context) {
	var req CreateMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	clientID, ok := normalizeClientID(c, req.Meal.ClientID)
	if !ok {
		return
	}
	meal := req.Meal.toModel(clientID)
	mealID, err := services.CreateMeal(c.Request.Context(), &meal)
	if err != nil {
		respondErr(c, err)
		return
	}
	if req.Log {
		day, err := services.FindMealPlanDay(c.Request.Context(), clock.ZerodTime(req.Offset))
		if err != nil {
			respondErr(c, err)
			return
		}
		if day == nil {
			apierr.NotFound(c, "Day not found")
			return
		}
		if err := services.CreateDayMeal(c.Request.Context(), &models.DayLog{
			DayID:  day.ID,
			MealID: mealID,
		}); err != nil {
			respondErr(c, err)
			return
		}
	}
	if req.Log && req.SaveToLibrary {
		sm := savedMealFromMealTemplate(&meal)
		if _, err := services.CreateSavedMeal(c.Request.Context(), sm); err != nil {
			respondErr(c, err)
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"meal_id": mealID})
}

type LogPlannedMealRequest struct {
	MealID uint `json:"meal_id"`
	// DayID is the diet day row id to log on. 0 (legacy clients) means today.
	DayID uint `json:"day_id"`
}

func PostLogPlanned(c *gin.Context) {
	var req LogPlannedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	day, err := services.FindMealPlanDayByRowIDOrToday(c.Request.Context(), req.DayID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "Day not found")
			return
		}
		respondErr(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	if err := services.LogPlannedMeal(c.Request.Context(), day.ID, req.MealID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "Planned meal not found on this day")
			return
		}
		respondErr(c, err)
		return
	}
	result, err := services.ReloadDayWithTotals(c.Request.Context(), day)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

type EditLoggedMealRequest struct {
	Meal                mealBody `json:"meal"`
	OldMealID           uint     `json:"oldMealID"`
	PlannedSourceMealID uint     `json:"planned_source_meal_id"`
	DayID               uint     `json:"day_id"`
}

func PostLogEdited(c *gin.Context) {
	var req EditLoggedMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	clientID, ok := normalizeClientID(c, req.Meal.ClientID)
	if !ok {
		return
	}
	meal := req.Meal.toModel(clientID)
	newMealID, err := services.CreateMeal(c.Request.Context(), &meal)
	if err != nil {
		respondErr(c, err)
		return
	}
	day, err := services.FindMealPlanDayByRowIDOrToday(c.Request.Context(), req.DayID)
	if err != nil {
		respondErr(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	if err := services.CreateDayMeal(c.Request.Context(), &models.DayLog{
		DayID:  day.ID,
		MealID: newMealID,
	}); err != nil {
		respondErr(c, err)
		return
	}
	if req.PlannedSourceMealID != 0 {
		if err := services.SetPlannedMealLogged(c.Request.Context(), day.ID, req.PlannedSourceMealID); err != nil {
			respondErr(c, err)
			return
		}
	}
	result, err := services.ReloadDayWithTotals(c.Request.Context(), day)
	if err != nil {
		respondErr(c, err)
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
	day, err := services.FindMealPlanDayByRowIDOrToday(c.Request.Context(), req.DayID)
	if err != nil {
		respondErr(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	clientID, ok := normalizeClientID(c, req.Meal.ClientID)
	if !ok {
		return
	}
	meal := req.Meal.toModel(clientID)
	if err := services.EditLoggedMeal(c.Request.Context(), day.ID, req.OldMealID, &meal); err != nil {
		respondErr(c, err)
		return
	}
	result, err := services.ReloadDayWithTotals(c.Request.Context(), day)
	if err != nil {
		respondErr(c, err)
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
	day, err := services.FindMealPlanDayByRowIDOrToday(c.Request.Context(), req.DayID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierr.NotFound(c, "Day not found")
			return
		}
		respondErr(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	if err := services.DeleteLoggedMeal(c.Request.Context(), day.ID, req.MealID); err != nil {
		respondErr(c, err)
		return
	}
	result, err := services.ReloadDayWithTotals(c.Request.Context(), day)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

type AddPlannedFromSavedRequest struct {
	SavedMealID uint   `json:"saved_meal_id"`
	Offset      int    `json:"offset"`
	ClientID    string `json:"client_id"`
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
	clientID, ok := normalizeClientID(c, req.ClientID)
	if !ok {
		return
	}
	if err := services.AddPlannedMealFromSavedMeal(c.Request.Context(), req.Offset, req.SavedMealID, clientID); err != nil {
		respondErr(c, err)
		return
	}
	day, err := services.FindMealPlanDay(c.Request.Context(), clock.ZerodTime(req.Offset))
	if err != nil {
		respondErr(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	result, err := services.ReloadDayWithTotals(c.Request.Context(), day)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}

type CopyPlannedDayRequest struct {
	FromOffset int   `json:"from_offset"`
	ToOffsets  []int `json:"to_offsets"`
}

// PostPlannedCopyDay appends the source day's planned meals to each target day.
func PostPlannedCopyDay(c *gin.Context) {
	var req CopyPlannedDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, err.Error())
		return
	}
	if len(req.ToOffsets) == 0 {
		apierr.BadRequest(c, "to_offsets is required")
		return
	}
	if len(req.ToOffsets) > 62 {
		apierr.BadRequest(c, "at most 62 target days")
		return
	}
	n, err := services.CopyPlannedMeals(c.Request.Context(), req.FromOffset, req.ToOffsets)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"copied_meals": n, "days": len(req.ToOffsets)})
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
	if err := services.ReorderPlannedMeals(c.Request.Context(), req.Offset, req.PlannedMealIDs); err != nil {
		respondErr(c, err)
		return
	}
	day, err := services.FindMealPlanDay(c.Request.Context(), clock.ZerodTime(req.Offset))
	if err != nil {
		respondErr(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	result, err := services.ReloadDayWithTotals(c.Request.Context(), day)
	if err != nil {
		respondErr(c, err)
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
	if err := services.DeletePlannedMeal(c.Request.Context(), req.Offset, req.PlannedMealID); err != nil {
		respondErr(c, err)
		return
	}
	day, err := services.FindMealPlanDay(c.Request.Context(), clock.ZerodTime(req.Offset))
	if err != nil {
		respondErr(c, err)
		return
	}
	if day == nil {
		apierr.NotFound(c, "Day not found")
		return
	}
	result, err := services.ReloadDayWithTotals(c.Request.Context(), day)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, dayWithTotalsResponse(result))
}
