package api

import (
	"be-simpletracker/database"
	"io"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitAPI() {
	f, _ := os.Create("api/gin.log")

	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)	 

	router := gin.Default()

	// router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
	// 	// Custom format
	// 	return fmt.Sprintf("[GIN] %s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
	// 	param.TimeStamp.Format(time.RFC1123),
	// 	param.ClientIP,
	// 	param.Method,
	// 	param.Path,
	// 	param.Request.Proto,
	// 	param.StatusCode,
	// 	param.Latency,
	// 	param.Request.UserAgent(),
	// 	param.ErrorMessage,
	// 	)
	// }))

	router.Use(BenchmarkMiddleware(router))

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
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


	router.Run("127.0.0.1:8080")
}