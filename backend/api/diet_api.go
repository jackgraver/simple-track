package api

import (
	"be-simpletracker/db/models"
	"be-simpletracker/db/services"
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
    // models.NewMealPlanModel(db)
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
        group.GET("/meal/all", f.getAllMeals)
        group.POST("/meal/log-planned", f.postLogPlanned)
        group.POST("/meal/logedited", NotImplemented)
    }
}

func (f *MealPlanFeature) getMealPlanToday(c *gin.Context) {
    day, daysErr := services.MealPlanToday(f.db)
    if daysErr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": daysErr.Error()})
        return
    }
    
    totalCalories := float32(0)
    totalProtein := float32(0)
    totalFiber := float32(0)
    for _, log := range day.Logs {
        for _, item := range log.Meal.Items {
            totalCalories += item.Food.Calories * item.Amount
            totalProtein += item.Food.Protein * item.Amount
            totalFiber += item.Food.Fiber * item.Amount
        }
    }

    c.JSON(http.StatusOK, gin.H{
		"day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
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

    start := startOfMonth.AddDate(0, 0, -int(startOfMonth.Weekday()))
    end := endOfMonth.AddDate(0, 0, 7-int(endOfMonth.Weekday()))

    // data, err := services.MealPlanRange(f.db, today, start, end)
    //TODO do business logic for grouping totals here
    data, err := services.ObjectRange[*models.Day](f.db, start, end)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "days":  data,
        "today": today,
        "range": gin.H{
            "start": start,
            "end":   end,
        },
        "month": target.Month(),
        "offset": offset,
    })
}

func (f *MealPlanFeature) getMealPlanDay(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    data, err := services.MealPlanDayByID(f.db, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if data == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
        return
    }

    c.JSON(http.StatusOK, data)
}

func (f *MealPlanFeature) getGoalsToday(c *gin.Context) {
    data, err := services.GoalsToday(f.db)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

func (f *MealPlanFeature) getAllFoods(c *gin.Context) {
    time.Sleep(3 * time.Second)
    data, err := services.AllFoods(f.db)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

func (f *MealPlanFeature) getAllMeals(c *gin.Context) {
    data, err := services.AllMeals(f.db)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}


type LogPlannedMealRequest struct {
    MealID uint `json:"meal_id"`
}

func (f *MealPlanFeature) postLogPlanned(c *gin.Context) {
    var req LogPlannedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err) // log to console
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    day, derr := services.FindMealPlanDay(f.db, utils.ZerodTime())
    if derr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": derr.Error()})
        return
    }

    if day == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
        return
    }

    meal, err:= services.MealByID(f.db, req.MealID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    error := services.CreateDayMeal(f.db, &models.DayLog{
        DayID: day.ID,
        MealID: meal.ID,
    })

    if error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": error.Error()})
        return
    }

    error = services.SetPlannedMealLogged(f.db, day.ID, meal.ID)
    if error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": error.Error()})
        return
    }

    day, err = services.MealPlanDayByID(f.db, int(day.ID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    fmt.Println(day.ID)
    c.JSON(http.StatusOK, gin.H{
        "day": day,
    })
}