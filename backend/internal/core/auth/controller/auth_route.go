package controller

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"be-simpletracker/internal/core/auth/models"
	authrepo "be-simpletracker/internal/core/auth/repository"
	"be-simpletracker/internal/core/auth/services"
	"be-simpletracker/internal/env"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CookieConfig struct {
	Name     string
	MaxAge   int
	Secure   bool
	SameSite http.SameSite
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,max=128"`
	Password string `json:"password" binding:"required,max=256"`
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

	result, err := service.Register(c.Request.Context(), services.RegisterInput{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidUsername), errors.Is(err, services.ErrPasswordTooShort), errors.Is(err, services.ErrPasswordTooLong):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

// LoginRequest: max lengths only bound the body; a password over 72 bytes simply fails
// (services.PasswordMaxBytes).
type LoginRequest struct {
	Username string `json:"username" binding:"required,max=128"`
	Password string `json:"password" binding:"required,max=256"`
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
	// Whether the username exists is only looked up when something is logged, and only the
	// existing names are logged verbatim.
	logAttempt := func(outcome string) {
		_, err := authrepo.FindUserByUsername(c.Request.Context(), req.Username)
		protection.LogAttempt(c.Request.Context(), outcome, clientIP, req.Username, err == nil)
	}
	decision := protection.Allow(clientIP, req.Username)
	if !decision.Allowed {
		if decision.LogRejection {
			outcome := "rate_limited"
			if decision.Reason == "credential_spray" {
				outcome = "credential_spray_blocked"
			}
			logAttempt(outcome)
			services.RecordRejection(c.Request.Context(), models.AuthEvent{Event: models.EventLoginRateLimited, ClientIP: clientIP, Detail: decision.Reason})
		}
		respondLoginRateLimited(c, decision.RetryAfter)
		return
	}

	protection.Wait(c.Request.Context(), decision.Delay)
	if c.Request.Context().Err() != nil {
		// The client went away during the backoff: don't check the password or log a failure.
		c.Abort()
		return
	}
	result, err := service.Login(c.Request.Context(), services.LoginInput{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials):
			recordLoginFailure(c, clientIP, err)
			sprayDecision := protection.RecordFailure(clientIP, req.Username)
			if !sprayDecision.Allowed {
				if sprayDecision.LogRejection {
					logAttempt("credential_spray_blocked")
				}
				respondLoginRateLimited(c, sprayDecision.RetryAfter)
				return
			}
			logAttempt("invalid_credentials")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		case errors.Is(err, services.ErrTokenGeneration):
			logAttempt("server_error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		default:
			logAttempt("server_error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	if err := services.RecordEvent(c.Request.Context(), models.AuthEvent{Event: models.EventLoginOK, UserID: &result.User.ID, ClientIP: clientIP}); err != nil {
		log.Printf("[auth] audit login_ok: %v", err)
	}
	setAuthCookie(c, cookie, result.Token)
	protection.RecordSuccess(clientIP, result.User.Username)
	protection.LogAttempt(c.Request.Context(), "success", clientIP, result.User.Username, true)
	c.JSON(http.StatusOK, AuthResponse{
		User:        result.User,
		Username:    result.User.Username,
		Environment: currentEnvironment(),
	})
}

// recordLoginFailure audits a failed login: the account if the username exists, never the
// attempted name (BACKLOG #19).
func recordLoginFailure(c *gin.Context, clientIP string, err error) {
	ev := models.AuthEvent{Event: models.EventLoginFailed, ClientIP: clientIP, Detail: "wrong_password"}
	var failure *services.LoginFailure
	if errors.As(err, &failure) {
		ev.Detail = failure.Reason
		if failure.UserID != 0 {
			id := failure.UserID
			ev.UserID = &id
		}
	}
	services.RecordRejection(c.Request.Context(), ev)
}

// GetCurrentUser answers /auth/me for userID (resolved by AuthMiddleware).
func GetCurrentUser(c *gin.Context, service *services.AuthService, userID uint) {
	setAuthResponseHeaders(c)
	user, err := service.CurrentUser(c.Request.Context(), userID)
	if err != nil {
		respondUserLookupError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user":        user,
		"username":    user.Username,
		"environment": currentEnvironment(),
	})
}

type updateProfileRequest struct {
	BirthYear *int `json:"birth_year"`
}

func UpdateCurrentUser(c *gin.Context, service *services.AuthService, userID uint) {
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.BirthYear == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "birth_year is required"})
		return
	}
	user, err := service.UpdateBirthYear(c.Request.Context(), userID, *req.BirthYear)
	if errors.Is(err, services.ErrInvalidBirthYear) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		respondUserLookupError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// respondUserLookupError: the user vanished after the middleware checked it (404), or the
// database failed (503, so the client doesn't treat it as signed out).
func respondUserLookupError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Could not load user"})
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
