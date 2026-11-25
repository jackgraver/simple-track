//go:build !e2e

// Package e2e is the test-only backend surface for the Playwright suite (e2e/README.md). The real
// implementation is compiled only with `-tags e2e`; this stub is what every other build gets.
package e2e

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Enabled reports whether this binary was built with the e2e tag.
const Enabled = false

// Install does nothing outside the e2e build.
func Install(*gin.Engine, *gorm.DB) {}
