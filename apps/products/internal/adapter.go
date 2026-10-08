package products

import (
	"context"

	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/geometry-software/gin-products/apps/providers/shared/mongoorm"
)

// StockPort deducts product stock through the providers service.
type StockPort interface {
	Deduct(context.Context, models.StockRequest) error
}
type stockHTTPAdapter struct{ baseURL string }

// Deduct applies an idempotent stock deduction.
func (a stockHTTPAdapter) Deduct(ctx context.Context, request models.StockRequest) error {
	return http.Call(ctx, "POST", a.baseURL+"/internal/providers/stock/deduct", "products", request, nil)
}

// New selects the product repository and stock endpoint for this process.
func New() *Service {
	baseURL := config.URL("adapters")
	return &Service{
		Products: mongoorm.RemoteRepository[models.Product]{BaseURL: baseURL, Caller: "products", Collection: "products"},
		Stock:    stockHTTPAdapter{baseURL: config.URL("providers")},
	}
}
