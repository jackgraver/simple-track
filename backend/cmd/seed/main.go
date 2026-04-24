// Standalone command: seeds the workout database (drops workout tables first - destructive).
// Usage: go run ./cmd/seed   (from backend/)
// Set DATABASE_URL to match the API server if not using defaults.
package main

import (
	dietmodels "be-simpletracker/internal/core/diet/models"
	"be-simpletracker/internal/core/tracking/models"
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
		&dietmodels.Plan{},
		&dietmodels.DietDay{},
		&dietmodels.Meal{},
		&dietmodels.MealItem{},
		&dietmodels.SavedMeal{},
		&dietmodels.SavedMealItem{},
		&dietmodels.PlannedMeal{},
		&dietmodels.DayLog{},
		&dietmodels.Food{},
		&dietmodels.CompositeFood{},
		&dietmodels.CompositeFoodItem{},
		&models.StepLog{},
		&models.BodyWeightLog{},
	); err != nil {
		fmt.Fprintf(os.Stderr, "AutoMigrate: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Workout seed completed.")
}
