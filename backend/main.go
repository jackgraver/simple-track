package main

import (
	"be-simpletracker/db"
	"be-simpletracker/handlers"
	"be-simpletracker/mealplanner"
	"be-simpletracker/workout"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := db.ConnectToDB()

	h := handlers.NewHandlers(db)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // or env var
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.SetTrustedProxies(nil)

	mealplanner.SetEndpoints(router, h)
	workout.SetEndpoints(router, h)

	router.GET("/api/foods", h.GetFoods)
	router.POST("/api/foods", h.AddFood)

	router.GET("/api/meals", h.GetMeals)
	router.POST("/api/meals", h.AddMeal)

	router.GET("/api/daily-goals", h.GetDailyGoals)

	router.GET("/api/today-meal-plan", h.GetTodayMealPlan)
	router.GET("/api/meal-plan-days", h.GetMealPlanDays)

	router.Run("127.0.0.1:8080")
}
