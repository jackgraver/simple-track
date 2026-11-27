package auth

import (
	"be-simpletracker/internal/core/auth/controller"
	"be-simpletracker/internal/core/auth/services"
	"be-simpletracker/internal/env"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	group := router.Group("/")
	registerEnabled := !env.IsProduction() && env.StringOr("REGISTER_ENABLED", "false") == "true"
	service := services.NewAuthService(GenerateToken)
	loginProtection := controller.NewLoginProtection(controller.LoginProtectionConfigFromEnv(), nil)
	cookie := controller.CookieConfig{
		Name:     AuthTokenCookieName,
		MaxAge:   CookieMaxAgeSeconds(),
		Secure:   CookieSecure(),
		SameSite: CookieSameSite(),
	}

	auth := group.Group("/auth")
	{
		if registerEnabled {
			auth.POST("/register", func(c *gin.Context) {
				controller.Register(c, service, cookie)
			})
		}
		auth.POST("/login", func(c *gin.Context) {
			controller.Login(c, service, cookie, loginProtection)
		})
		auth.POST("/logout", func(c *gin.Context) {
			controller.Logout(c, cookie)
		})
		auth.GET("/me", AuthMiddleware(db), func(c *gin.Context) {
			controller.GetCurrentUser(c, service, MustUser(c).ID)
		})
		auth.PATCH("/me", AuthMiddleware(db), func(c *gin.Context) {
			controller.UpdateCurrentUser(c, service, MustUser(c).ID)
		})
	}
}
