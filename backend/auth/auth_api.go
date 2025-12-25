package auth

import (
	"be-simpletracker/auth/models"
	"be-simpletracker/auth/routes"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) RegisterRoutes(router *gin.Engine) {
	group := router.Group("/")
	routes.RegisterAuthRoutes(group, h.db, AuthMiddleware(), GenerateToken)
}

func (h *Handler) Migrate() error {
	return h.db.AutoMigrate(&models.User{})
}

