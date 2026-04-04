// Standalone command: seeds the workout database (drops workout tables first - destructive).
// Usage: go run ./cmd/seed   (from backend/)
// Set DATABASE_URL to match the API server if not using defaults.
package main

import (
	"be-simpletracker/internal/auth/models"
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

	db.AutoMigrate(&models.User{})

	fmt.Println("Workout seed completed.")
}
