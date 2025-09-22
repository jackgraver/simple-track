package handlers

import (
	"net/http"
	"time"

	"be-simpletracker/db"
	"be-simpletracker/db/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	DB *gorm.DB
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{DB: db}
}

// Food handlers
func (h *Handlers) GetFoods(c *gin.Context) {
	foods, err := db.GetFoods(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, foods)
}

func (h *Handlers) AddFood(c *gin.Context) {
	var foodData struct {
		Name     string  	`json:"name" binding:"required"`
		Unit     string  	`json:"unit" binding:"required"`
		Calories float32 	`json:"calories" binding:"required"`
		Protein  float32 	`json:"protein"`
		Fiber    float32 	`json:"fiber"`
	}

	if err := c.ShouldBindJSON(&foodData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	food, err := db.AddFood(h.DB, foodData.Name, foodData.Unit, foodData.Calories, foodData.Protein, foodData.Fiber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, food)
}

// Meal handlers
func (h *Handlers) GetMeals(c *gin.Context) {
	meals, err := db.GetMeals(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, meals)
}

func (h *Handlers) AddMeal(c *gin.Context) {
	var mealData struct {
		Name  string             `json:"name" binding:"required"`
		Items []models.MealItem `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&mealData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	meal, err := db.AddMeal(h.DB, mealData.Name, mealData.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, meal)
}

func (h *Handlers) GetDailyGoals(c *gin.Context) {
	dateStr := c.Query("date")
	var date time.Time
	var err error

	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	} else {
		date = time.Now()
	}

	goals, err := db.GetDailyGoals(h.DB, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, goals)
}

// Health check
func (h *Handlers) Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"status":  "healthy",
	})
}


func (h *Handlers) GetMealPlanDays(c *gin.Context) {
	days, err := db.GetMealPlanDays(h.DB)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"days": days,
		"today": time.Now(),
	})
}

func (h *Handlers) GetTodayMealPlan(c *gin.Context) {
    date := time.Now()
    days, err := db.GetTodayMealPlan(h.DB, date)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "date": date, // or any other format
        "days": days,
    })
}

func (h *Handlers) GetMealNames(c *gin.Context) {
	names, err := db.GetMealNames(h.DB, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, names)
}