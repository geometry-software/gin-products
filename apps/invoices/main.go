package main

import (
	"context"
	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	domain "github.com/geometry-software/gin-products/apps/invoices/internal"
	"github.com/gin-gonic/gin"
	"log/slog"
	"os"
)

func main() {
	if err := config.Load(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	if len(config.Get("APP_JWT_SECRET", "")) < 32 {
		slog.Error("APP_JWT_SECRET must contain at least 32 characters")
		os.Exit(1)
	}
	r := http.New("invoices")
	domain.Routes(r, domain.New())
	r.GET("/readyz", func(c *gin.Context) {
		for _, service := range []string{"adapters", "login"} {
			if err := http.Call(c.Request.Context(), "GET", config.URL(service)+"/readyz", "invoices", nil, nil); err != nil {
				controller.Error(c, err)
				return
			}
		}
		c.JSON(200, gin.H{"status": "ready"})
	})
	http.Run("invoices", r, func(context.Context) {})
}
