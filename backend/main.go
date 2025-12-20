package main

import (
	"be-simpletracker/database"
	diet "be-simpletracker/features/diet"
	workout "be-simpletracker/features/workout"
	"be-simpletracker/utils"
	"io"
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
	f, err := os.Create("gin.log")
	if err != nil {
		// If we can't create the log file, just use stdout
		gin.DefaultWriter = os.Stdout
	} else {
		gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	}	 

	router := gin.Default()

	router.Use(utils.BenchmarkMiddleware(router))

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
	
	CreateFeatures(db, router)

	addr := getEnv("LISTEN_ADDR", "0.0.0.0:8080")
	router.Run(addr)
}

func CreateFeatures(db *gorm.DB, router *gin.Engine) {
	diet := diet.NewHandler(db)
	diet.RegisterRoutes(router)

	workout := workout.NewHandler(db)
	workout.RegisterRoutes(router)
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
