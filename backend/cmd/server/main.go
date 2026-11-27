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
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	// Contexts derived from a *gin.Context (context.WithTimeout(c, ...)) still reach the
	// request context, so auth.UserID works on them.
	router.ContextWithFallback = true

	// First, so every response (including CORS preflights and 404s) gets it.
	router.Use(hardening(defaultBodyLimit, bodyLimits))
	router.Use(corsMiddleware(env.SliceOr("CORS_ORIGINS", ",", nil)))

	if !env.IsProduction() {
		router.Use(utils.BenchmarkMiddleware(router))
	}

	trustedProxies := env.SliceOr("TRUSTED_PROXIES", ",", nil)
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		log.Fatalf("config: invalid TRUSTED_PROXIES: %v", err)
	}
	if env.IsProduction() && len(trustedProxies) == 0 {
		log.Printf("config: TRUSTED_PROXIES is empty; login protection will use the direct peer IP")
	}
	db, err := database.ConnectApp()
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	if err := checkSchema(db); err != nil {
		log.Fatalf("database schema: %v", err)
	}
	if err := checkRole(db); err != nil {
		if env.IsProduction() {
			log.Fatalf("database role: %v", err)
		}
		log.Printf("database role: warning: %v", err)
	}

	// Test-only /__e2e endpoints; a no-op unless built with -tags e2e (e2e/README.md).
	e2e.Install(router, db)
	// After e2e.Install (its holds must not sit in an open transaction), before every route that
	// reaches the database (docs/multi-user/DECISIONS.md, "B2").
	router.Use(database.Middleware(db))
	CreateFeatures(db, router)
	// Last, after the schema check: see healthHandler.
	router.GET("/health", healthHandler(db))

	addr := env.StringOr("LISTEN_ADDR", "0.0.0.0:8080")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// After the first signal, restore default handling so a second Ctrl-C/SIGTERM kills at once.
	go func() { <-ctx.Done(); stop() }()
	if err := serve(ctx, newHTTPServer(addr, router), db); err != nil {
		log.Fatal(err)
	}
}

// checkRole: the server must not connect as a role that row-level security doesn't apply to.
func checkRole(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return database.CheckAppRole(ctx, sqlDB)
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
	auth.RegisterRoutes(router, db)

	authMW := auth.AuthMiddleware(db)

	diet.RegisterRoutes(router, authMW)

	workout.RegisterRoutes(router, authMW)

	trackingHandler := tracking.NewHandler()
	trackingHandler.RegisterRoutes(router, authMW)

	moneyHandler := money.NewHandler()
	moneyHandler.RegisterRoutes(router, authMW)

	mediaHandler := media.NewHandler()
	mediaHandler.RegisterRoutes(router, authMW)
}
