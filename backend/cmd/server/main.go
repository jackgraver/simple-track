package main

import (
	"be-simpletracker/internal/core/auth"
	diet "be-simpletracker/internal/core/diet"
	tracking "be-simpletracker/internal/core/tracking"
	workout "be-simpletracker/internal/core/workout"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/utils"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// Set the mode to release mode (stops DEBUG logging like all defined routes)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(utils.BenchmarkMiddleware(router))

	corsOrigins := getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000,http://192.168.4.78:3000,http://192.168.4.64:3000")
	origins := splitString(corsOrigins, ",")

	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Set-Cookie"},
		MaxAge:           12 * time.Hour,
	}))

	router.SetTrustedProxies(nil)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	db, err := database.ConnectToPostgres()
	if err != nil {
		panic(err)
	}

	CreateFeatures(db, router)

	addr := getEnv("LISTEN_ADDR", "0.0.0.0:8080")
	fmt.Println("Server is running on port", addr)
	router.Run(addr)
}

func CreateFeatures(db *gorm.DB, router *gin.Engine) {
	authHandler := auth.NewHandler(db)
	authHandler.RegisterRoutes(router)

	authMW := auth.AuthMiddleware()
	dietHandler := diet.NewHandler(db)
	if err := dietHandler.Migrate(); err != nil {
		panic(err)
	}
	dietHandler.RegisterRoutes(router, authMW)

	workoutHandler := workout.NewHandler(db)
	workoutHandler.RegisterRoutes(router, authMW)

	trackingHandler := tracking.NewHandler(db)
	if err := trackingHandler.Migrate(); err != nil {
		panic(err)
	}
	trackingHandler.RegisterRoutes(router, authMW)
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
