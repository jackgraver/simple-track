package api

import (
	"be-simpletracker/database"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func InitAPI() {
	f, _ := os.Create("api/gin.log")

	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)	 

	router := gin.Default()

	router.Use(BenchmarkMiddleware(router))

	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000")
	origins := []string{}
	for _, origin := range splitString(corsOrigins, ",") {
		origins = append(origins, origin)
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.SetTrustedProxies(nil)

	db, err := database.ConnectToSqlite()
	if err != nil {
		panic(err)
	}
	database.DefineRoutes(router)
	
	diet := NewMealPlanFeature(db)
	diet.SetEndpoints(router)

	workout := NewWorkoutFeature(db)
	workout.SetEndpoints(router)


	addr := getEnv("LISTEN_ADDR", "0.0.0.0:8080")
	router.Run(addr)
}

func splitString(s, sep string) []string {
	var result []string
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}