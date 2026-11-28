package auth

import (
	"net/http"

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
	// Setup links and password changes have their own limiter, so their attempts never use up a
	// source's sign-in budget (family members share one home IP).
	accountProtection := controller.NewLoginProtection(controller.LoginProtectionConfigFromEnv(), nil)
	cookie := controller.CookieConfig{
		Name:     AuthTokenCookieName,
		MaxAge:   CookieMaxAgeSeconds(),
		Secure:   CookieSecure(),
		SameSite: CookieSameSite(),
	}

	auth := group.Group("/auth")
	{
		if registerEnabled {
			auth.POST("/register", OutsideRequestTx(db), func(c *gin.Context) {
				controller.Register(c, service, cookie)
			})
		}
		auth.POST("/login", OutsideRequestTx(db), func(c *gin.Context) {
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
		auth.POST("/password", AuthMiddleware(db), func(c *gin.Context) {
			controller.ChangePassword(c, service, cookie, accountProtection, actor(c))
		})
		// Setup links: no session yet. They run in the request transaction (the token's row is
		// visible only through app.setup_token_hash) and hash the new password before it begins.
		auth.POST("/setup/inspect", func(c *gin.Context) {
			controller.InspectSetupLink(c, service)
		})
		auth.POST("/setup", func(c *gin.Context) {
			controller.CompleteSetup(c, service, cookie, accountProtection)
		})
	}

	admin := group.Group("/admin", AuthMiddleware(db), RequireAdmin)
	{
		admin.GET("/users", func(c *gin.Context) { controller.ListUsers(c, service) })
		admin.POST("/users", func(c *gin.Context) { controller.CreateUser(c, service, actor(c)) })
		admin.POST("/users/:id/setup-link", func(c *gin.Context) { controller.CreateSetupLink(c, service, actor(c)) })
		admin.POST("/users/:id/disable", func(c *gin.Context) { controller.SetUserDisabled(c, service, actor(c), true) })
		admin.POST("/users/:id/enable", func(c *gin.Context) { controller.SetUserDisabled(c, service, actor(c), false) })
		admin.GET("/auth-events", func(c *gin.Context) { controller.ListAuthEvents(c, service) })
	}
}

// actor is the signed-in user and client IP, for the audit log.
func actor(c *gin.Context) services.Actor {
	return services.Actor{UserID: MustUser(c).ID, ClientIP: c.ClientIP()}
}

// RequireAdmin answers 403 unless the signed-in user is the admin. Use after AuthMiddleware.
func RequireAdmin(c *gin.Context) {
	if user, ok := CurrentUser(c); !ok || !user.IsAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin only"})
		return
	}
	c.Next()
}
