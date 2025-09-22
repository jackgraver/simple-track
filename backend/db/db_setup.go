package db

import (
	"be-simpletracker/db/models"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectToDB initializes the database connection and runs auto-migration
func ConnectToDB() *gorm.DB {
	dsn := "host=localhost user=postgres password=pass123 dbname=postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database" + err.Error())
	}

	DropAllTables(db)
	// Auto-migrate schema
	db.AutoMigrate(&models.Food{}, &models.Meal{}, &models.MealItem{}, &models.DayGoals{}, &models.MealPlanDay{}, &models.DayMeal{})
	
	// Seed database with starter data
	seedDatabase(db)
	
	return db
}

// seedDatabase adds predefined starter data to the database
func seedDatabase(db *gorm.DB) {
	chicken := models.Food{Name: "Chicken Breast", Unit: "g", Calories: 1.65, Protein: 0.31, Fiber: 0}
	rice    := models.Food{Name: "White Rice", Unit: "g", Calories: 1.30, Protein: 0.02, Fiber: 0.01}
	yogurt  := models.Food{Name: "Greek Yogurt", Unit: "g", Calories: 0.59, Protein: 0.10, Fiber: 0}
	berries := models.Food{Name: "Blueberries", Unit: "g", Calories: 0.57, Protein: 0.01, Fiber: 0.025}

	db.Create(&chicken)
	db.Create(&rice)
	db.Create(&yogurt)
	db.Create(&berries)
	
	chickenRice := models.Meal{
		Name: "Chicken & Rice Bowl",
		Items: []models.MealItem{
			{FoodID: chicken.ID, Amount: 200},
			{FoodID: rice.ID, Amount: 100},
		},
	}
	proteinBreakfast := models.Meal{
		Name: "Protein Breakfast",
		Items: []models.MealItem{
			{FoodID: chicken.ID, Amount: 200},
			{FoodID: rice.ID, Amount: 100},
		},
	}
	greekYogurt := models.Meal{
		Name: "Greek Yogurt Bowl",
		Items: []models.MealItem{
			{FoodID: chicken.ID, Amount: 200},
			{FoodID: rice.ID, Amount: 100},
		},
	}
	
	db.Create(&chickenRice)
	db.Create(&proteinBreakfast)
	db.Create(&greekYogurt)

	// 2. Create day with meals and items
	day1 := models.MealPlanDay{
		Date: time.Date(2025, time.September, 20, 0, 0, 0, 0, time.UTC),
		Meals: []models.DayMeal{
			{MealID: chickenRice.ID, Status: "expected"},
			{MealID: chickenRice.ID, Status: "actual"},
		},
		Goals: models.DayGoals{
			Calories: 2000,
			Protein:  150,
			Fiber:    65,
		},
	}
	db.Create(&day1)
	
	day2 := models.MealPlanDay{
		Date: time.Date(2025, time.September, 21, 0, 0, 0, 0, time.UTC),
		Meals: []models.DayMeal{
			{MealID: proteinBreakfast.ID, Status: "expected"},
			{MealID: proteinBreakfast.ID, Status: "actual"},
			{MealID: greekYogurt.ID, Status: "expected"},
		},
		Goals: models.DayGoals{
			Calories: 1900,
			Protein:  160,
			Fiber:    45,
		},
	}
	db.Create(&day2)

	year := 2025
	month := time.September
	daysInMonth := 30 // September has 30 days

	for day := 1; day <= daysInMonth; day++ {
		if day == 20 || day == 21 {
			continue
		}

		date := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

		mpd := models.MealPlanDay{
			Date: date,
			Goals: models.DayGoals{
				Calories: 2000,
				Protein:  150,
				Fiber:    50,
			},
		}

		if err := db.Create(&mpd).Error; err != nil {
			fmt.Printf("Failed to create day %v: %v\n", date, err)
		}
	}
}

func DropAllTables(db *gorm.DB) error {
    return db.Migrator().DropTable(&models.Food{}, &models.Meal{}, &models.MealItem{}, &models.DayGoals{}, &models.MealPlanDay{}, &models.DayMeal{})
}