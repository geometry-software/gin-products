package invoices

import (
	"context"
	"net/url"

	"github.com/geometry-software/gin-products/apps/providers/shared/config"
	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/geometry-software/gin-products/apps/providers/shared/mongoorm"
)

type productsHTTPAdapter struct{ baseURL string }

type usersHTTPAdapter struct{ baseURL string }

// Get returns a record by ID.
func (a usersHTTPAdapter) Get(ctx context.Context, id string) (out models.User, err error) {
	err = http.Call(ctx, "GET", a.baseURL+"/api/users/"+url.PathEscape(id), "", nil, &out)
	return
}

// Exchange asks Login to verify a Firebase ID token and resolve its user.
func (a usersHTTPAdapter) Exchange(ctx context.Context, idToken string) (out models.User, err error) {
	err = http.Call(ctx, "POST", a.baseURL+"/api/auth/session", "", map[string]string{"idToken": idToken}, &out)
	return
}

// Resolve returns records for an internal service request.
func (a productsHTTPAdapter) Resolve(ctx context.Context, ids []string) (out []models.Product, err error) {
	err = http.Call(ctx, "POST", a.baseURL+"/internal/products/resolve", "invoices", map[string]any{"ids": ids}, &out)
	return
}

// Deduct applies an idempotent stock deduction.
func (a productsHTTPAdapter) Deduct(ctx context.Context, request models.StockRequest) error {
	return http.Call(ctx, "POST", a.baseURL+"/internal/products/stock/deduct", "invoices", request, nil)
}

// New wires invoice persistence and the Products HTTP integration.
func New() *Service {
	return &Service{
		Invoices: mongoorm.RemoteRepository[models.Invoice]{BaseURL: config.URL("adapters"), Caller: "invoices", Collection: "invoices"},
		Products: productsHTTPAdapter{baseURL: config.URL("products")},
		Users:    usersHTTPAdapter{baseURL: config.URL("login")},
	}
}
