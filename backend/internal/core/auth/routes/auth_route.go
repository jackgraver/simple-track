package routes

import (
	"be-simpletracker/internal/core/auth/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// authCookieName matches auth.AuthTokenCookieName — routes cannot import the auth package without a cycle with auth.RegisterRoutes callers.
const authCookieName = "auth_token"

type AuthHandler struct {
	db             *gorm.DB
	generateToken  func(string) (string, error)
	cookieMaxAge   int
	cookieSecure   bool
	cookieSameSite http.SameSite
}

func NewAuthHandler(db *gorm.DB, generateToken func(string) (string, error), cookieMaxAge int, cookieSecure bool, cookieSameSite http.SameSite) *AuthHandler {
	return &AuthHandler{
		db:             db,
		generateToken:  generateToken,
		cookieMaxAge:   cookieMaxAge,
		cookieSecure:   cookieSecure,
		cookieSameSite: cookieSameSite,
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token    string      `json:"token"`
	User     models.User `json:"user"`
	Username string      `json:"username"`
}

// Register creates a new user account
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := h.generateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.Password = ""

	c.SetSameSite(h.cookieSameSite)
	c.SetCookie(authCookieName, token, h.cookieMaxAge, "/", "", h.cookieSecure, true)

	c.JSON(http.StatusCreated, AuthResponse{
		Token:    token,
		User:     user,
		Username: user.Username,
	})
}

// Login authenticates a user and returns a JWT token
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	token, err := h.generateToken(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	user.Password = ""

	c.SetSameSite(h.cookieSameSite)
	c.SetCookie(authCookieName, token, h.cookieMaxAge, "/", "", h.cookieSecure, true)

	c.JSON(http.StatusOK, AuthResponse{
		Token:    token,
		User:     user,
		Username: user.Username,
	})
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	usernameStr := username.(string)

	var user models.User
	if err := h.db.Where("username = ?", usernameStr).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"user": user, "username": user.Username})
}

// Logout clears the authentication cookie
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetSameSite(h.cookieSameSite)
	c.SetCookie(authCookieName, "", -1, "/", "", h.cookieSecure, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func RegisterAuthRoutes(group *gin.RouterGroup, db *gorm.DB, authMiddleware gin.HandlerFunc, generateToken func(string) (string, error), cookieMaxAge int, cookieSecure bool, cookieSameSite http.SameSite) {
	handler := NewAuthHandler(db, generateToken, cookieMaxAge, cookieSecure, cookieSameSite)

	auth := group.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/logout", handler.Logout)
		auth.GET("/me", authMiddleware, handler.GetCurrentUser)
	}
}
