package invoices

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/geometry-software/gin-products/apps/providers/shared/mongoorm"
)

// ProductsPort provides product lookup and stock deduction for invoices.
type ProductsPort interface {
	Resolve(context.Context, []string) ([]models.Product, error)
	Deduct(context.Context, models.StockRequest) error
}

// UsersPort looks up users from the Login service.
type UsersPort interface {
	Get(context.Context, string) (models.User, error)
	Exchange(context.Context, string) (models.User, error)
}

// Service holds the repositories and integrations for this domain.
type Service struct {
	Invoices mongoorm.Repository[models.Invoice]
	Products ProductsPort
	Users    UsersPort
}

// TokenResult contains an invoice-only bearer token and its user.
type TokenResult struct {
	AccessToken string      `json:"accessToken"`
	TokenType   string      `json:"tokenType"`
	ExpiresAt   time.Time   `json:"expiresAt"`
	User        models.User `json:"user"`
}

// List returns a paginated collection.
func (s *Service) List(ctx context.Context, query url.Values) (models.Page[models.Invoice], error) {
	return s.Invoices.List(ctx, query)
}

// Get returns a record by ID.
func (s *Service) Get(ctx context.Context, id string) (models.Invoice, error) {
	return s.Invoices.Get(ctx, id)
}

// Resolve returns records for an internal service request.
func (s *Service) Resolve(ctx context.Context, ids []string) ([]models.Invoice, error) {
	invoices := make([]models.Invoice, 0, len(ids))
	for _, id := range ids {
		invoice, err := s.Invoices.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

// Input defines the validated request body for this domain.
type Input struct {
	Name        string            `json:"name" binding:"required,max=200"`
	Description string            `json:"description" binding:"max=4000"`
	Items       []models.Quantity `json:"items" binding:"required,min=1,max=100,dive"`
	Version     int64             `json:"version"`
}

// Snapshot builds an invoice snapshot from product records.
func Snapshot(q Input, products []models.Product) (models.Invoice, error) {
	items, err := mongoorm.Normalize(q.Items)
	if err != nil {
		return models.Invoice{}, err
	}
	catalog := map[string]models.Product{}
	for _, p := range products {
		catalog[p.ID] = p
	}
	invoice := models.Invoice{Name: q.Name, Description: q.Description, Status: "pending", Items: []models.InvoiceItem{}}
	for _, item := range items {
		p, ok := catalog[item.ProductID]
		if !ok || !p.Active || p.Quantity < item.Quantity {
			return invoice, http.Fail(409, "Product unavailable or insufficient stock")
		}
		if p.Price < 0 || p.Price > 1000000000 {
			return invoice, http.Fail(422, "Invalid product price")
		}
		invoice.Items = append(invoice.Items, models.InvoiceItem{Quantity: item, Name: p.Name, Description: p.Description, UnitPrice: p.Price})
		invoice.Total += p.Price * item.Quantity
	}
	return invoice, nil
}
func (s *Service) prepare(ctx context.Context, q Input) (models.Invoice, error) {
	ids := make([]string, 0, len(q.Items))
	for _, item := range q.Items {
		ids = append(ids, item.ProductID)
	}
	products, err := s.Products.Resolve(ctx, ids)
	if err != nil {
		return models.Invoice{}, err
	}
	return Snapshot(q, products)
}

// Create validates and persists a new record.
func (s *Service) Create(ctx context.Context, q Input) (models.Invoice, error) {
	v, err := s.prepare(ctx, q)
	if err != nil {
		return v, err
	}
	return s.Invoices.Create(ctx, v)
}

// Update replaces a record while checking its current version.
func (s *Service) Update(ctx context.Context, id string, q Input) (models.Invoice, error) {
	old, err := s.Invoices.Get(ctx, id)
	if err != nil {
		return old, err
	}
	if old.Status != "pending" || old.Version != q.Version {
		return old, http.Fail(409, "Only a current pending invoice can be edited")
	}
	next, err := s.prepare(ctx, q)
	if err != nil {
		return next, err
	}
	next.Entity = old.Entity
	return s.Invoices.Replace(ctx, id, next)
}

// Reject marks a pending invoice as rejected.
func (s *Service) Reject(ctx context.Context, id string) (models.Invoice, error) {
	v, err := s.Invoices.Get(ctx, id)
	if err != nil {
		return v, err
	}
	if v.Status == "rejected" {
		return v, nil
	}
	if v.Status != "pending" {
		return v, http.Fail(409, "Only pending invoices can be rejected")
	}
	v.Status = "rejected"
	return s.Invoices.Replace(ctx, id, v)
}

// Confirm deducts stock and completes an invoice.
func (s *Service) Confirm(ctx context.Context, id string) (models.Invoice, error) {
	v, err := s.Invoices.Get(ctx, id)
	if err != nil {
		return v, err
	}
	if v.Status == "complete" {
		return v, nil
	}
	if v.Status == "rejected" {
		return v, http.Fail(409, "Rejected invoice cannot be confirmed")
	}
	if v.Status == "pending" {
		v.Status = "confirming"
		v, err = s.Invoices.Replace(ctx, id, v)
		if err != nil {
			return v, err
		}
	}
	if v.Status != "confirming" {
		return v, http.Fail(409, "Invalid invoice state")
	}
	items := make([]models.Quantity, 0, len(v.Items))
	for _, item := range v.Items {
		items = append(items, item.Quantity)
	}
	err = s.Products.Deduct(ctx, models.StockRequest{OperationID: v.ID, Items: items})
	if err != nil {
		// Only a definitive stock conflict is safe to roll back. Timeouts leave the
		// invoice locked for a retry with the same idempotency key and snapshot.
		var e *http.Error
		if errors.As(err, &e) && e.Status == 409 {
			v.Status = "pending"
			_, _ = s.Invoices.Replace(ctx, id, v)
		}
		return v, err
	}
	v.Status = "complete"
	out, err := s.Invoices.Replace(ctx, id, v)
	if err != nil {
		current, readErr := s.Invoices.Get(ctx, id)
		if readErr == nil && current.Status == "complete" {
			return current, nil
		}
	}
	return out, err
}
