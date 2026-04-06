// Standalone command: seeds the workout database (drops workout tables first - destructive).
// Usage: go run ./cmd/seed   (from backend/)
// Set DATABASE_URL to match the API server if not using defaults.
package main

import (
	// "be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/core/diet/models"
	"be-simpletracker/internal/database"
	"fmt"
	"os"
)

func main() {
	db, err := database.ConnectToPostgres()
	if err != nil {
		fmt.Fprintf(os.Stderr, "database: %v\n", err)
		os.Exit(1)
	}
	// if err := workoutseed.Run(db); err != nil {
	// 	fmt.Fprintf(os.Stderr, "seed: %v\n", err)
	// 	os.Exit(1)
	// }

	// db.AutoMigrate(&models.User{})

	if err := db.AutoMigrate(
		&models.Plan{},
		&models.DietDay{},
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

	fmt.Println("Workout seed completed.")
}
