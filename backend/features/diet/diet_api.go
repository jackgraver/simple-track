package diet

import (
	"be-simpletracker/features/diet/models"
	"be-simpletracker/features/diet/routes"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

// NewHandler creates a new workout plan handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes registers all workout feature routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/diet")
	
	// seedDatabase(h.db)

	// Register plan sub-domain routes
	routes.RegisterDietPlanRoutes(group, h.db)
	routes.RegisterMealRoutes(group, h.db)
	routes.RegisterDietLogRoutes(group, h.db)
}

func seedDatabase(db *gorm.DB) error {
	fmt.Println("Migrating meal plan database")
	if err := db.Migrator().DropTable(
		&models.DayLog{},
		&models.PlannedMeal{},
		&models.MealItem{},
		&models.SavedMealItem{},
		&models.Meal{},
		&models.SavedMeal{},
		&models.Food{},
		&models.Day{},
		&models.Plan{},
	); err != nil {
		fmt.Printf("Failed to drop meal plan database: %v\n", err)
	}

	if err := db.AutoMigrate(
		&models.Plan{},
		&models.Day{},
		&models.Meal{},
		&models.MealItem{},
		&models.SavedMeal{},
		&models.SavedMealItem{},
		&models.PlannedMeal{},
		&models.DayLog{},
		&models.Food{},
	); err != nil {
		fmt.Printf("Failed to migrate meal plan database: %v\n", err)
	}

	fmt.Println("Seeding meal plan database")

	bulk := models.Plan{Name: "Bulk",
				Calories: 2400,
				Protein:  150,
				Fiber:    50,
				Carbs:    150,
			}
	db.Create(&bulk)

	year := 2025
	start := time.Date(year, time.September, 1, 0, 0, 0, 0, time.Local)
	end := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.Local)

	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		mpd := models.Day{
			Date: date,
			Plan: bulk,
			PlannedMeals: []models.PlannedMeal{},
		}

		if err := db.Create(&mpd).Error; err != nil {
			fmt.Printf("Failed to create day %v: %v\n", date, err)
		}
	}
	return nil;
}