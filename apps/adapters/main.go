package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	adapters "github.com/geometry-software/gin-products/apps/adapters/internal"
	adapterhttp "github.com/geometry-software/gin-products/apps/adapters/internal/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
)

func main() {
	if err := config.Load(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	adapter, err := adapters.New(ctx)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	r := adapterhttp.Router(adapter)
	http.Run("adapters", r, adapter.Close)
}
