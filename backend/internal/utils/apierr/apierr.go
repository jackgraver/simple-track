package apierr

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type body struct {
	Error string `json:"error"`
}

func write(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, body{Error: msg})
}

func BadRequest(c *gin.Context, msg string) { write(c, http.StatusBadRequest, msg) }

func NotFound(c *gin.Context, msg string) { write(c, http.StatusNotFound, msg) }

func Conflict(c *gin.Context, msg string) { write(c, http.StatusConflict, msg) }

func Unauthorized(c *gin.Context, msg string) { write(c, http.StatusUnauthorized, msg) }

func Forbidden(c *gin.Context, msg string) { write(c, http.StatusForbidden, msg) }

func Internal(c *gin.Context, err error) {
	slog.Error("internal server error",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"err", err,
	)
	write(c, http.StatusInternalServerError, "Internal server error")
}
