package products

import (
	"context"
	"net/url"

	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/geometry-software/gin-products/apps/providers/shared/mongoorm"
)

// Service holds the repositories and integrations for this domain.
type Service struct {
	Products mongoorm.Repository[models.Product]
	Stock    StockPort
}

// Create validates and persists a new record.
func (s *Service) Create(ctx context.Context, product models.Product) (models.Product, error) {
	product.Entity = models.Entity{}
	return s.Products.Create(ctx, product)
}

// List returns a paginated collection.
func (s *Service) List(ctx context.Context, query url.Values) (models.Page[models.Product], error) {
	return s.Products.List(ctx, query)
}

// Get returns a record by ID.
func (s *Service) Get(ctx context.Context, id string) (models.Product, error) {
	return s.Products.Get(ctx, id)
}

// Update replaces a record while checking its current version.
func (s *Service) Update(ctx context.Context, id string, product models.Product) (models.Product, error) {
	old, err := s.Products.Get(ctx, id)
	if err != nil {
		return old, err
	}
	if product.Version != old.Version {
		return old, http.Fail(409, "Supply the current version")
	}
	product.Entity = old.Entity
	return s.Products.Replace(ctx, id, product)
}

// Delete removes a record by its ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	product, err := s.Products.Get(ctx, id)
	if err != nil {
		return err
	}
	return s.Products.Delete(ctx, id, product.Version)
}

// Resolve returns records for an internal service request.
func (s *Service) Resolve(ctx context.Context, ids []string) ([]models.Product, error) {
	values := make([]models.Product, 0, len(ids))
	for _, id := range ids {
		v, err := s.Products.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}
	return values, nil
}

// Deduct applies an idempotent stock deduction.
func (s *Service) Deduct(ctx context.Context, req models.StockRequest) error {
	return s.Stock.Deduct(ctx, req)
}
