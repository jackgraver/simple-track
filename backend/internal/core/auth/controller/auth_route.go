package controller

import (
	"errors"
	"net/http"

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
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

func Register(c *gin.Context, service *services.AuthService, cookie CookieConfig) {
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
		Token:       result.Token,
		User:        result.User,
		Username:    result.User.Username,
		Environment: currentEnvironment(),
	})
}

type AuthResponse struct {
	Token       string      `json:"token"`
	User        models.User `json:"user"`
	Username    string      `json:"username"`
	Environment string      `json:"environment"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context, service *services.AuthService, cookie CookieConfig) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := service.Login(services.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		case errors.Is(err, services.ErrTokenGeneration):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	setAuthCookie(c, cookie, result.Token)
	c.JSON(http.StatusOK, AuthResponse{
		Token:       result.Token,
		User:        result.User,
		Username:    result.User.Username,
		Environment: currentEnvironment(),
	})
}

func GetCurrentUser(c *gin.Context, service *services.AuthService) {
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

func currentEnvironment() string {
	return env.StringOr("APP_ENV", env.StringOr("GO_ENV", "development"))
}
