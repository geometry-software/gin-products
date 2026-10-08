package shipping

import (
	"context"

	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/geometry-software/gin-products/apps/providers/shared/mongoorm"
)

type invoicesHTTPAdapter struct{ baseURL string }

// Resolve returns records for an internal service request.
func (a invoicesHTTPAdapter) Resolve(ctx context.Context, ids []string) (out []models.Invoice, err error) {
	err = http.Call(ctx, "POST", a.baseURL+"/internal/invoices/resolve", "shipping", map[string]any{"ids": ids}, &out)
	return
}

// New wires shipment persistence to the adapter process and the owning services.
func New() *Service {
	return &Service{
		Shipments: mongoorm.RemoteRepository[models.Shipment]{BaseURL: config.URL("adapters"), Caller: "shipping", Collection: "shipments"},
		Invoices:  invoicesHTTPAdapter{baseURL: config.URL("invoices")},
	}
}
