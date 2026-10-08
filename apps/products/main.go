package main

import (
	"context"
	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	domain "github.com/geometry-software/gin-products/apps/products/internal"
	"github.com/gin-gonic/gin"
	"log/slog"
	"os"
)

func main() {
	if err := config.Load(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	r := http.New("products")
	domain.Routes(r, domain.New())
	r.GET("/readyz", func(c *gin.Context) {
		for _, service := range []string{"adapters", "providers"} {
			if err := http.Call(c.Request.Context(), "GET", config.URL(service)+"/readyz", "products", nil, nil); err != nil {
				controller.Error(c, err)
				return
			}
		}
		c.JSON(200, gin.H{"status": "ready"})
	})
	http.Run("products", r, func(context.Context) {})
}
