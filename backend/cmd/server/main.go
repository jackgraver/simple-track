package main

import (
	"be-simpletracker/internal/core/auth"
	diet "be-simpletracker/internal/core/diet"
	media "be-simpletracker/internal/core/media"
	money "be-simpletracker/internal/core/money"
	tracking "be-simpletracker/internal/core/tracking"
	workout "be-simpletracker/internal/core/workout"
	"be-simpletracker/internal/database"
	"be-simpletracker/internal/e2e"
	"be-simpletracker/internal/env"
	"be-simpletracker/internal/migrate"
	"be-simpletracker/internal/utils"
	"be-simpletracker/migrations"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	if err := env.Load(); err != nil {
		log.Fatal(err)
	}
	if _, err := env.String("JWT_SECRET"); err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := auth.ValidateProductionConfig(); err != nil {
		log.Fatalf("config: %v", err)
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	if !env.IsProduction() {
		router.Use(utils.BenchmarkMiddleware(router))
	}

	defaultCORS := "http://localhost:5173,http://localhost:3000,http://192.168.4.78:3000,http://192.168.4.64:3000"
	corsOrigins := env.StringOr("CORS_ORIGINS", defaultCORS)
	origins := splitString(corsOrigins, ",")

	router.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Set-Cookie"},
		MaxAge:           12 * time.Hour,
	}))

	trustedProxies := env.SliceOr("TRUSTED_PROXIES", ",", nil)
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		log.Fatalf("config: invalid TRUSTED_PROXIES: %v", err)
	}
	if env.IsProduction() && len(trustedProxies) == 0 {
		log.Printf("config: TRUSTED_PROXIES is empty; login protection will use the direct peer IP")
	}
	db, err := database.ConnectToPostgres()
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := checkSchema(db); err != nil {
		log.Fatalf("database schema: %v", err)
	}

	// Test-only /__e2e endpoints; a no-op unless built with -tags e2e (e2e/README.md).
	e2e.Install(router, db)
	CreateFeatures(db, router)
	// Last, after the schema check: see healthHandler.
	router.GET("/health", healthHandler(db))

	addr := env.StringOr("LISTEN_ADDR", "0.0.0.0:8080")
	fmt.Println("Server is running on port", addr)
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

// checkSchema refuses to start against a database that is missing migrations
// this binary was built with. Migrations are applied separately by
// cmd/migrate (deploy step, or the `migrate` dev service).
func checkSchema(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	m, err := migrate.New(sqlDB, migrations.FS, log.Printf)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Check(ctx)
}

func CreateFeatures(db *gorm.DB, router *gin.Engine) {
	auth.RegisterRoutes(router)

	authMW := auth.AuthMiddleware()

	diet.RegisterRoutes(router, authMW)

	workout.RegisterRoutes(router, authMW)

	trackingHandler := tracking.NewHandler(db)
	trackingHandler.RegisterRoutes(router, authMW)

	moneyHandler := money.NewHandler(db)
	moneyHandler.RegisterRoutes(router, authMW)

	mediaHandler := media.NewHandler(db)
	mediaHandler.RegisterRoutes(router, authMW)
}
