package api

import (
	"be-simpletracker/database/models"
	"be-simpletracker/database/services"
	"be-simpletracker/utils"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MealPlanFeature struct {
	BaseFeature[models.MealPlanModel]
}

func NewMealPlanFeature(db *gorm.DB) *MealPlanFeature {
    models.NewMealPlanModel(db)
    var feature = models.NewMealPlanModel(db)
    feature.MigrateDatabase()

	return &MealPlanFeature{
		BaseFeature[models.MealPlanModel]{
			db: db,
		},
	}
}

func (f *MealPlanFeature) SetEndpoints(router *gin.Engine) {
    group := router.Group("/mealplan") 
    {
        group.GET("/today", f.getMealPlanToday)
        group.GET("/week", f.getMealPlanWeek)
        group.GET("/month", f.getMealPlanMonth)
        group.GET("/day/:id" , f.getMealPlanDay)
        group.GET("/goals/today", f.getGoalsToday)
        group.GET("/food/all", f.getAllFoods)
        group.POST("/food/add", f.postAddFood)
        group.GET("/meal/all", f.getAllMeals)
        group.GET("/meal/:id", f.getMeal)
        group.POST("/meal/new", f.postNewMeal)
        group.POST("/meal/log-planned", f.postLogPlanned)
        group.POST("/meal/logedited", f.postLogEdited)
        group.POST("/meal/editlogged", f.postEditLogged)
        group.DELETE("/meal/logged", f.deleteLoggedMeal)
        group.GET("/plan/all", f.getAllPlans)
    }
}

func (f *MealPlanFeature) getMealPlanToday(c *gin.Context) {
    offsetStr := c.Query("offset")
    offset, _ := strconv.Atoi(offsetStr)

    day, daysErr := services.MealPlanToday(f.db, offset)
    if daysErr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": daysErr.Error()})
        return
    }
    
    totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(f.db, day.ID)

    c.JSON(http.StatusOK, gin.H{
		"day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
		"today": time.Now(),
	})
}

func (f *MealPlanFeature) getMealPlanWeek(c *gin.Context) {
    today := time.Now()
    start := today.AddDate(0, 0, -3) // 3 days before
	end := today.AddDate(0, 0, 3)    // 3 days after
    // data, err := services.MealPlanRange(f.db, today, start, end)
    data, err := services.ObjectRange[*models.Day](f.db, start, end)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
		"days": data,
		"today": time.Now(),
	})
}

func (f *MealPlanFeature) getMealPlanMonth(c *gin.Context) {
    offsetStr := c.Query("monthoffset")
    offset, _ := strconv.Atoi(offsetStr)

    today := time.Now()
    target := today.AddDate(0, offset, 0)

    startOfMonth := time.Date(target.Year(), target.Month(), 1, 0, 0, 0, 0, target.Location())
    endOfMonth := startOfMonth.AddDate(0, 1, -1)

    data, err := services.ObjectRange[*models.Day](f.db, startOfMonth, endOfMonth)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
		"days": data,
		"today": time.Now(),
		"range": gin.H{
			"start": startOfMonth,
			"end": endOfMonth,
		},
		"month": target.Month(),
		"offset": offset,
	})
}

func (f *MealPlanFeature) getMealPlanDay(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    day, err := services.MealPlanDayByID(f.db, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(f.db, day.ID)

    c.JSON(http.StatusOK, gin.H{
        "day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
    })
}

func (f *MealPlanFeature) getGoalsToday(c *gin.Context) {
    goals, err := services.GoalsToday(f.db)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, goals)
}

func (f *MealPlanFeature) getAllFoods(c *gin.Context) {
    excludeIDsStr := c.Query("exclude")
    var excludeIDs []uint
    if excludeIDsStr != "" {
        ids := []uint{}
        for _, idStr := range []string{excludeIDsStr} {
            if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
                ids = append(ids, uint(id))
            }
        }
        excludeIDs = ids
    }

    foods, err := services.AllFoods(f.db, excludeIDs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "foods": foods,
    })
}

func (f *MealPlanFeature) postAddFood(c *gin.Context) {
    var food models.Food
    if err := c.BindJSON(&food); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    createdFood, err := services.CreateFood(f.db, &food)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, createdFood)
}

func (f *MealPlanFeature) getAllMeals(c *gin.Context) {
    excludeIDsStr := c.Query("exclude")
    var excludeIDs []uint
    if excludeIDsStr != "" {
        ids := []uint{}
        for _, idStr := range []string{excludeIDsStr} {
            if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
                ids = append(ids, uint(id))
            }
        }
        excludeIDs = ids
    }

    meals, err := services.AllMeals(f.db, excludeIDs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "meals": meals,
    })
}

func (f *MealPlanFeature) getMeal(c *gin.Context) {
    idStr := c.Param("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    meal, err := services.MealByID(f.db, uint(id))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, meal)
}

func (f *MealPlanFeature) postNewMeal(c *gin.Context) {
    var meal models.Meal
    if err := c.BindJSON(&meal); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    mealID, err := services.CreateMeal(f.db, &meal)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "id": mealID,
        "meal": meal,
    })
}

func (f *MealPlanFeature) postLogPlanned(c *gin.Context) {
    var req LogPlannedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    day, err := services.FindMealPlanDay(f.db, utils.ZerodTime(0))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if day == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Day not found"})
        return
    }

    err = services.SetPlannedMealLogged(f.db, day.ID, req.MealID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    day, err = services.MealPlanDayByID(f.db, int(day.ID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(f.db, day.ID)

    c.JSON(http.StatusOK, gin.H{
        "day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
    })
}

func (f *MealPlanFeature) postLogEdited(c *gin.Context) {
    var req EditLoggedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    day, err := services.FindMealPlanDay(f.db, utils.ZerodTime(0))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if day == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Day not found"})
        return
    }

    err = services.UpdateDayLogMeal(f.db, day.ID, req.OldMealID, req.Meal.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    day, err = services.MealPlanDayByID(f.db, int(day.ID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(f.db, day.ID)

    c.JSON(http.StatusOK, gin.H{
        "day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
    })
}

func (f *MealPlanFeature) postEditLogged(c *gin.Context) {
    var req EditLoggedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    day, err := services.FindMealPlanDay(f.db, utils.ZerodTime(0))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if day == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Day not found"})
        return
    }

    err = services.UpdateDayLogMeal(f.db, day.ID, req.OldMealID, req.Meal.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    day, err = services.MealPlanDayByID(f.db, int(day.ID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(f.db, day.ID)

    c.JSON(http.StatusOK, gin.H{
        "day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
    })
}

func (f *MealPlanFeature) deleteLoggedMeal(c *gin.Context) {
    var req DeleteLoggedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    day, err := services.FindMealPlanDay(f.db, utils.ZerodTime(0))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    err = services.DeleteLoggedMeal(f.db, day.ID, req.MealID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    day, err = services.MealPlanDayByID(f.db, int(day.ID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    totalCalories, totalProtein, totalFiber, totalCarbs := services.CalculateTotals(f.db, day.ID)

    c.JSON(http.StatusOK, gin.H{
        "day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
    })
}

// getAllPlans retrieves all plans with support for pagination, sorting, and filtering
// Uses the Plan service which encapsulates repository logic
func (f *MealPlanFeature) getAllPlans(c *gin.Context) {
	ctx := c.Request.Context()
	
	// Use the service function - encapsulates all repository logic
	// result, err := services.GetAllPlans(ctx, f.db, c)
    result, err := services.GetAllWithOptions[*models.Plan](ctx, f.db, c, "id", true)
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

type EditLoggedMealRequest struct {
    Meal models.Meal `json:"meal"`
    OldMealID uint `json:"oldMealID"`
}

type LogPlannedMealRequest struct {
    MealID uint `json:"mealID"`
}

type DeleteLoggedMealRequest struct {
    MealID uint `json:"mealID"`
}
