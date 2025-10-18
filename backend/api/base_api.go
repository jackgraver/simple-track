package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BaseFeature[T any] struct {
	db *gorm.DB
	SetEndpoints func()
}

func NotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not Implemented"})
}