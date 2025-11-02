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
    excludeParam := c.QueryArray("exclude") // e.g. /mealplan/food/all?exclude=1&exclude=3
    var excludeIDs []uint
    for _, idStr := range excludeParam {
        if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
            excludeIDs = append(excludeIDs, uint(id))
        }
    }

    data, err := services.AllFoods(f.db, excludeIDs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"foods": data})
}

type AddFoodRequest struct {
    Food models.Food `json:"food"`
}

func (f *MealPlanFeature) postAddFood(c *gin.Context) {
    var req AddFoodRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    food, err := services.CreateFood(f.db, &req.Food)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"food": food})
}

func (f *MealPlanFeature) getAllMeals(c *gin.Context) {
    time.Sleep(2000 * time.Millisecond)
    excludeParam := c.QueryArray("exclude") // e.g. /mealplan/food/all?exclude=1&exclude=3
    var excludeIDs []uint
    for _, idStr := range excludeParam {
        if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
            excludeIDs = append(excludeIDs, uint(id))
        }
    }

    data, err := services.AllMeals(f.db, excludeIDs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, data)
}

func (f *MealPlanFeature) getMeal(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    data, err := services.MealByID(f.db, uint(id))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if data == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
        return
    }    
    c.JSON(http.StatusOK, gin.H{"meal": data})
}

type NewMealRequest struct {
    Meal models.Meal `json:"meal"`
    Log bool `json:"log"`
}

func (f *MealPlanFeature) postNewMeal(c *gin.Context) {
    var req NewMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    meal, err := services.CreateMeal(f.db, &req.Meal)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    if req.Log {
        day, derr := services.FindMealPlanDay(f.db, utils.ZerodTime(0))
        if derr != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": derr.Error()})
            return
        }

        err = services.CreateDayMeal(f.db, &models.DayLog{
            DayID: day.ID,
            MealID: meal,
        })
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"meal_id": meal})
}


type LogPlannedMealRequest struct {
    MealID uint `json:"meal_id"`
}

func (f *MealPlanFeature) postLogPlanned(c *gin.Context) {
    var req LogPlannedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    meal, err:= services.MealByID(f.db, req.MealID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    f.postLogMeal(c, meal.ID, true)
}

type LogEditedMealRequest struct {
    Meal models.Meal `json:"meal"`
}

func (f *MealPlanFeature) postLogEdited(c *gin.Context) {
    //bind body
    var req LogEditedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    //create new meal
    newMealID, err := services.CreateMeal(f.db, &req.Meal)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    f.postLogMeal(c, newMealID, false)
}

func (f *MealPlanFeature) postLogMeal(c *gin.Context, mealID uint, setLogged bool) {
    day, derr := services.FindMealPlanDay(f.db, utils.ZerodTime(0))
    if derr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": derr.Error()})
        return
    }

    if day == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Not Found"})
        return
    }
    error := services.CreateDayMeal(f.db, &models.DayLog{
        DayID: day.ID,
        MealID: mealID,
    })

    if error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": error.Error()})
        return
    }

    if setLogged {
        error = services.SetPlannedMealLogged(f.db, day.ID, mealID)
        if error != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": error.Error()})
            return
        }
    }
   
    day, err := services.MealPlanDayByID(f.db, int(day.ID))
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

type DeleteLoggedMealRequest struct {
    MealID uint `json:"meal_id"`
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

func (f *MealPlanFeature) getAllPlans(c *gin.Context) {
    plans, err := services.GetAllPlans(f.db)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "plans": plans,
    })
}

type EditLoggedMealRequest struct {
    Meal models.Meal `json:"meal"`
    OldMealID uint `json:"oldMealID"`
}

func (f *MealPlanFeature) postEditLogged(c *gin.Context) {
    var req EditLoggedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    //create new meal
    newMealID, err := services.CreateMeal(f.db, &req.Meal)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    //get day
    day, err := services.FindMealPlanDay(f.db, utils.ZerodTime(0))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    //update logged meal
    err = services.UpdateDayLogMeal(f.db, day.ID, req.OldMealID, newMealID)
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