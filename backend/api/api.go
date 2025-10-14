package api

import (
	"be-simpletracker/db"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitAPI() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.SetTrustedProxies(nil)

	db, err := db.ConnectToSqlite()
	if err != nil {
		panic(err)
	}

	diet := NewMealPlanFeature(db)
	diet.SetEndpoints(router)

	workout := NewWorkoutFeature(db)
	workout.SetEndpoints(router)

	router.Run("127.0.0.1:8080")
}