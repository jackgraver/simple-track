package routes

import (
	"be-simpletracker/database/services"
	"be-simpletracker/diet/models"
	"be-simpletracker/utils"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MealHandler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewMealHandler(db *gorm.DB) *MealHandler {
	return &MealHandler{db: db}
}

func RegisterMealRoutes(group *gin.RouterGroup, db *gorm.DB) {
	h := NewMealHandler(db)

	meals := group.Group("/meals")
	{
		meals.GET("/food/all", h.getAllFoods)
        meals.POST("/food/add", h.postAddFood)
        meals.GET("/meal/all", h.getAllMeals)
        meals.GET("/meal/:id", h.getMeal)
        meals.POST("/meal/new", h.postNewMeal)
        meals.POST("/meal/log-planned", h.postLogPlanned)
        meals.POST("/meal/logedited", h.postLogEdited)
        meals.POST("/meal/editlogged", h.postEditLogged)
        meals.DELETE("/meal/logged", h.deleteLoggedMeal)
	}
}

func (h *MealHandler) getAllFoods(c *gin.Context) {
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

    foods, err := services.AllFoods(h.db, excludeIDs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "foods": foods,
    })
}

func (h *MealHandler) postAddFood(c *gin.Context) {
    var food models.Food
    if err := c.BindJSON(&food); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    createdFood, err := services.CreateFood(h.db, &food)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, createdFood)
}

func (h *MealHandler) getAllMeals(c *gin.Context) {
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

    meals, err := services.AllMeals(h.db, excludeIDs)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "meals": meals,
    })
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

func (h *MealHandler) postNewMeal(c *gin.Context) {
    var meal models.Meal
    if err := c.BindJSON(&meal); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    mealID, err := services.CreateMeal(h.db, &meal)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "id": mealID,
        "meal": meal,
    })
}


type LogPlannedMealRequest struct {
    MealID uint `json:"meal_id"`
}

func (h *MealHandler) postLogPlanned(c *gin.Context) {
    var req LogPlannedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
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

    err = services.SetPlannedMealLogged(h.db, day.ID, req.MealID)
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
        "day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
    })
}

type EditLoggedMealRequest struct {
    Meal models.Meal `json:"meal"`
    OldMealID uint `json:"oldMealID"`
}

func (h *MealHandler) postLogEdited(c *gin.Context) {
    var req EditLoggedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
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

    err = services.UpdateDayLogMeal(h.db, day.ID, req.OldMealID, req.Meal.ID)
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
        "day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
    })
}

func (h *MealHandler) postEditLogged(c *gin.Context) {
    var req EditLoggedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
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

    err = services.UpdateDayLogMeal(h.db, day.ID, req.OldMealID, req.Meal.ID)
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

func (h *MealHandler) deleteLoggedMeal(c *gin.Context) {
    var req DeleteLoggedMealRequest
    if err := c.BindJSON(&req); err != nil {
        fmt.Println("BindJSON error:", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    day, err := services.FindMealPlanDay(h.db, utils.ZerodTime(0))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    err = services.DeleteLoggedMeal(h.db, day.ID, req.MealID)
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
        "day": day,
        "totalCalories": totalCalories,
        "totalProtein": totalProtein,
        "totalFiber": totalFiber,
        "totalCarbs": totalCarbs,
    })
}