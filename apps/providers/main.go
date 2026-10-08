package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/geometry-software/gin-products/apps/providers/shared/controller"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	providers "github.com/geometry-software/gin-products/apps/providers/internal"
	"github.com/gin-gonic/gin"
)

func main() {
	if err := config.Load(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	connection, err := providers.Open(ctx)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	r := http.New("providers")
	if err := providers.Register(r, connection.Database); err != nil {
		connection.Close(ctx)
		slog.Error(err.Error())
		os.Exit(1)
	}
	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()
		if err := connection.Ready(ctx); err != nil {
			controller.Error(c, err)
			return
		}
		c.JSON(200, gin.H{"status": "ready"})
	})
	http.Run("providers", r, connection.Close)
}
