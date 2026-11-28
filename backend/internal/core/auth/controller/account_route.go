package controller

import (
	"errors"
	"net/http"
	"strconv"

	"be-simpletracker/internal/core/auth/models"
	"be-simpletracker/internal/core/auth/services"
	"be-simpletracker/internal/utils/apierr"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Account handlers (lane C6): the admin's user management, setup links and the audit log, the
// public setup-link endpoints, and changing one's own password. Tokens travel in request bodies,
// never in API URLs, so they stay out of request logs.

// respondAccountError maps the account services' errors to responses.
func respondAccountError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidUsername), errors.Is(err, services.ErrPasswordTooShort),
		errors.Is(err, services.ErrPasswordTooLong), errors.Is(err, services.ErrSamePassword),
		errors.Is(err, services.ErrSelfDisable):
		apierr.BadRequest(c, err.Error())
	case errors.Is(err, services.ErrWrongPassword):
		// 400, not 401: the client treats 401 as "signed out".
		apierr.BadRequest(c, "Current password is incorrect")
	case errors.Is(err, services.ErrUsernameExists):
		apierr.Conflict(c, "Username already exists")
	case errors.Is(err, services.ErrUserDisabled):
		apierr.Conflict(c, "User is disabled; enable them first")
	case errors.Is(err, gorm.ErrRecordNotFound):
		apierr.NotFound(c, "User not found")
	case errors.Is(err, services.ErrSetupLinkInvalid):
		c.AbortWithStatusJSON(http.StatusGone, gin.H{"error": "This link has expired or was already used. Ask for a new one."})
	default:
		apierr.Internal(c, err)
	}
}

func pathUserID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		apierr.NotFound(c, "User not found")
		return 0, false
	}
	return uint(id), true
}

func ListUsers(c *gin.Context, service *services.AuthService) {
	users, err := service.ListUsers(c.Request.Context())
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

type createUserRequest struct {
	Username string `json:"username" binding:"required,max=64"`
}

func CreateUser(c *gin.Context, service *services.AuthService, actor services.Actor) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "username is required")
		return
	}
	user, link, err := service.CreateUser(c.Request.Context(), actor, req.Username)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user, "link": link})
}

func CreateSetupLink(c *gin.Context, service *services.AuthService, actor services.Actor) {
	id, ok := pathUserID(c)
	if !ok {
		return
	}
	link, err := service.CreateSetupLink(c.Request.Context(), actor, id)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"link": link})
}

func SetUserDisabled(c *gin.Context, service *services.AuthService, actor services.Actor, disabled bool) {
	id, ok := pathUserID(c)
	if !ok {
		return
	}
	user, err := service.SetDisabled(c.Request.Context(), actor, id, disabled)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func ListAuthEvents(c *gin.Context, service *services.AuthService) {
	before, _ := strconv.ParseUint(c.Query("before"), 10, 32)
	limit, _ := strconv.Atoi(c.Query("limit"))
	events, err := service.ListEvents(c.Request.Context(), uint(before), limit)
	if err != nil {
		apierr.Internal(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

type setupTokenRequest struct {
	Token string `json:"token" binding:"required,max=128"`
}

// InspectSetupLink tells the /setup page whose link it is, or 410 if it's dead.
func InspectSetupLink(c *gin.Context, service *services.AuthService) {
	setAuthResponseHeaders(c)
	var req setupTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondAccountError(c, services.ErrSetupLinkInvalid)
		return
	}
	info, err := service.InspectSetupLink(c.Request.Context(), req.Token)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	c.JSON(http.StatusOK, info)
}

type completeSetupRequest struct {
	Token    string `json:"token" binding:"required,max=128"`
	Password string `json:"password" binding:"required,max=256"`
}

// setupProtectionKey is the "username" setup attempts count against, on the account limiter
// (not the sign-in one): per source it allows MaxPerUsername attempts per window, and all
// sources together go through the username-wide backoff.
const setupProtectionKey = "setup-link"

// CompleteSetup sets the password through a setup link and signs the user in.
func CompleteSetup(c *gin.Context, service *services.AuthService, cookie CookieConfig, protection *LoginProtection) {
	setAuthResponseHeaders(c)
	var req completeSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "Invalid setup request")
		return
	}
	// A password that breaks the rules doesn't count as an attempt.
	if err := services.ValidatePassword(req.Password); err != nil {
		respondAccountError(c, err)
		return
	}
	clientIP := c.ClientIP()
	key := setupProtectionKey
	decision := protection.Allow(clientIP, key)
	if !decision.Allowed {
		if decision.LogRejection {
			services.RecordRejection(c.Request.Context(), models.AuthEvent{Event: models.EventLoginRateLimited, ClientIP: clientIP, Detail: "setup_link:" + decision.Reason})
		}
		respondLoginRateLimited(c, decision.RetryAfter)
		return
	}
	if !waitBackoff(c, protection, decision) {
		return
	}
	result, err := service.CompleteSetup(c.Request.Context(), services.Actor{ClientIP: clientIP}, req.Token, req.Password)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	protection.RecordSuccess(clientIP, key)
	setAuthCookie(c, cookie, result.Token)
	c.JSON(http.StatusOK, AuthResponse{User: result.User, Username: result.User.Username, Environment: currentEnvironment()})
}

// waitBackoff applies the limiter's username-wide backoff (S1) before a password is checked. It
// returns false if the client went away meanwhile (nothing checked, nothing logged).
func waitBackoff(c *gin.Context, protection *LoginProtection, decision LoginDecision) bool {
	protection.Wait(c.Request.Context(), decision.Delay)
	if c.Request.Context().Err() != nil {
		c.Abort()
		return false
	}
	return true
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,max=256"`
	NewPassword     string `json:"new_password" binding:"required,max=256"`
}

// ChangePassword changes the signed-in user's password. Other sessions end; this one gets a
// fresh cookie. Wrong current passwords are limited like logins, on their own counter.
func ChangePassword(c *gin.Context, service *services.AuthService, cookie CookieConfig, protection *LoginProtection, actor services.Actor) {
	setAuthResponseHeaders(c)
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierr.BadRequest(c, "Current and new password are required")
		return
	}
	if err := services.ValidatePassword(req.NewPassword); err != nil {
		respondAccountError(c, err)
		return
	}
	key := "password-change:" + strconv.FormatUint(uint64(actor.UserID), 10)
	decision := protection.Allow(actor.ClientIP, key)
	if !decision.Allowed {
		if decision.LogRejection {
			services.RecordRejection(c.Request.Context(), models.AuthEvent{Event: models.EventLoginRateLimited, UserID: &actor.UserID, ClientIP: actor.ClientIP, Detail: "password_change:" + decision.Reason})
		}
		respondLoginRateLimited(c, decision.RetryAfter)
		return
	}
	if !waitBackoff(c, protection, decision) {
		return
	}
	result, err := service.ChangePassword(c.Request.Context(), actor, req.CurrentPassword, req.NewPassword)
	if err != nil {
		respondAccountError(c, err)
		return
	}
	protection.RecordSuccess(actor.ClientIP, key)
	setAuthCookie(c, cookie, result.Token)
	c.JSON(http.StatusOK, gin.H{"user": result.User})
}
