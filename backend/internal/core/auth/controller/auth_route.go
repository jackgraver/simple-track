package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/core/auth/services"
	"be-simpletracker/internal/env"

	"github.com/gin-gonic/gin"
)

type CookieConfig struct {
	Name     string
	MaxAge   int
	Secure   bool
	SameSite http.SameSite
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,max=128"`
	Password string `json:"password" binding:"required,max=72"`
	Email    string `json:"email" binding:"max=254"`
}

func Register(c *gin.Context, service *services.AuthService, cookie CookieConfig) {
	setAuthResponseHeaders(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16*1024)
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := service.Register(services.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrUsernameExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		case errors.Is(err, services.ErrPasswordHash):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		case errors.Is(err, services.ErrUserCreation):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		case errors.Is(err, services.ErrTokenGeneration):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	setAuthCookie(c, cookie, result.Token)
	c.JSON(http.StatusCreated, AuthResponse{
		User:        result.User,
		Username:    result.User.Username,
		Environment: currentEnvironment(),
	})
}

type AuthResponse struct {
	User        models.User `json:"user"`
	Username    string      `json:"username"`
	Environment string      `json:"environment"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required,max=128"`
	Password string `json:"password" binding:"required,max=72"`
}

func Login(c *gin.Context, service *services.AuthService, cookie CookieConfig, protection *LoginProtection) {
	setAuthResponseHeaders(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16*1024)
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid login request"})
		return
	}

	clientIP := c.ClientIP()
	decision := protection.Allow(clientIP, req.Username)
	if !decision.Allowed {
		if decision.LogRejection {
			outcome := "rate_limited"
			if decision.Reason == "credential_spray" {
				outcome = "credential_spray_blocked"
			}
			protection.LogAttempt(c.Request.Context(), outcome, clientIP, req.Username)
		}
		respondLoginRateLimited(c, decision.RetryAfter)
		return
	}

	result, err := service.Login(services.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			sprayDecision := protection.RecordFailure(clientIP, req.Username)
			if !sprayDecision.Allowed {
				if sprayDecision.LogRejection {
					protection.LogAttempt(c.Request.Context(), "credential_spray_blocked", clientIP, req.Username)
				}
				respondLoginRateLimited(c, sprayDecision.RetryAfter)
				return
			}
			protection.LogAttempt(c.Request.Context(), "invalid_credentials", clientIP, req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		case errors.Is(err, services.ErrTokenGeneration):
			protection.LogAttempt(c.Request.Context(), "server_error", clientIP, req.Username)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		default:
			protection.LogAttempt(c.Request.Context(), "server_error", clientIP, req.Username)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	setAuthCookie(c, cookie, result.Token)
	protection.RecordSuccess(clientIP, result.User.Username)
	protection.LogAttempt(c.Request.Context(), "success", clientIP, result.User.Username)
	c.JSON(http.StatusOK, AuthResponse{
		User:        result.User,
		Username:    result.User.Username,
		Environment: currentEnvironment(),
	})
}

func GetCurrentUser(c *gin.Context, service *services.AuthService) {
	setAuthResponseHeaders(c)
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	usernameStr, ok := username.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	user, err := service.CurrentUser(usernameStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"username":    user.Username,
		"environment": currentEnvironment(),
	})
}

func Logout(c *gin.Context, cookie CookieConfig) {
	c.SetSameSite(cookie.SameSite)
	c.SetCookie(cookie.Name, "", -1, "/", "", cookie.Secure, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func setAuthCookie(c *gin.Context, cookie CookieConfig, token string) {
	c.SetSameSite(cookie.SameSite)
	c.SetCookie(cookie.Name, token, cookie.MaxAge, "/", "", cookie.Secure, true)
}

func setAuthResponseHeaders(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
}

func respondLoginRateLimited(c *gin.Context, retryAfter time.Duration) {
	retryAfterSeconds := int((retryAfter + time.Second - 1) / time.Second)
	if retryAfterSeconds < 1 {
		retryAfterSeconds = 1
	}
	c.Header("Retry-After", strconv.Itoa(retryAfterSeconds))
	c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts. Try again later"})
}

func currentEnvironment() string {
	return env.StringOr("APP_ENV", env.StringOr("GO_ENV", "development"))
}
