// Package controller centralizes JSON validation and HTTP error responses
// for the domain controllers.
package controller

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/gin-gonic/gin"
)

// Error writes a typed domain error or a safe response for unexpected failures.
func Error(c *gin.Context, err error) {
	var domain *http.Error
	if errors.As(err, &domain) {
		c.AbortWithStatusJSON(domain.Status, domain)
		return
	}
	slog.Error("request failed", "path", c.FullPath(), "type", fmt.Sprintf("%T", err))
	c.AbortWithStatusJSON(503, gin.H{"error": "Service temporarily unavailable"})
}

// BindJSON validates a request body and writes the shared 400 response on failure.
func BindJSON[T any](c *gin.Context) (T, bool) {
	var value T
	if !ValidateJSON(c, &value) {
		return value, false
	}
	return value, true
}

// ValidateJSON binds a request into an existing value and reports validation errors.
func ValidateJSON(c *gin.Context, value any) bool {
	if err := c.ShouldBindJSON(value); err != nil {
		Error(c, http.Fail(400, "Invalid request body"))
		return false
	}
	return true
}

// JSON runs a controller action and serializes its result or error.
func JSON[T any](c *gin.Context, status int, action func() (T, error)) {
	value, err := action()
	if err != nil {
		Error(c, err)
		return
	}
	c.JSON(status, value)
}

// NoContent runs a controller action and returns 204 on success.
func NoContent(c *gin.Context, action func() error) {
	if err := action(); err != nil {
		Error(c, err)
		return
	}
	c.Status(204)
}
