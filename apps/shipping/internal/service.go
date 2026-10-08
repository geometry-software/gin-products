package shipping

import (
	"context"
	"net/url"
	"time"

	"github.com/geometry-software/gin-products/apps/providers/shared/http"
	"github.com/geometry-software/gin-products/apps/providers/shared/models"
	"github.com/geometry-software/gin-products/apps/providers/shared/mongoorm"
	"github.com/google/uuid"
)

// InvoicesPort resolves completed invoice snapshots for shipping.
type InvoicesPort interface {
	Resolve(context.Context, []string) ([]models.Invoice, error)
}

// Service holds the repositories and integrations for this domain.
type Service struct {
	Shipments mongoorm.Repository[models.Shipment]
	Invoices  InvoicesPort
}

// List returns a paginated collection.
func (s *Service) List(ctx context.Context, query url.Values) (models.Page[models.Shipment], error) {
	return s.Shipments.List(ctx, query)
}

// Get returns a record by ID.
func (s *Service) Get(ctx context.Context, id string) (models.Shipment, error) {
	return s.Shipments.Get(ctx, id)
}

// Input defines the validated request body for this domain.
type Input struct {
	Recipient  models.Recipient `json:"recipient" binding:"required"`
	InvoiceIDs []string         `json:"invoiceIds" binding:"required,min=1,max=100,unique,dive,uuid"`
}

// Create validates and persists a new record.
func (s *Service) Create(ctx context.Context, q Input) (models.Shipment, error) {
	invoices, err := s.Invoices.Resolve(ctx, q.InvoiceIDs)
	if err != nil {
		return models.Shipment{}, err
	}
	if len(invoices) != len(q.InvoiceIDs) {
		return models.Shipment{}, http.Fail(409, "Invoices missing")
	}
	for _, v := range invoices {
		if v.Status != "complete" {
			return models.Shipment{}, http.Fail(409, "Shipping requires completed invoices")
		}
	}
	shipment := models.Shipment{TrackingNumber: "GIN-" + uuid.NewString(), Status: "created", Recipient: q.Recipient, Invoices: invoices, TrackingEvents: []models.TrackingEvent{{Status: "created", OccurredAt: time.Now().UTC()}}}
	return s.Shipments.Create(ctx, shipment)
}

// CanTransition reports whether a shipment status change is allowed.
func CanTransition(from, to string) bool {
	allowed := map[string][]string{"created": {"in_transit", "cancelled"}, "in_transit": {"out_for_delivery", "exception"}, "out_for_delivery": {"delivered", "exception"}, "exception": {"in_transit", "cancelled"}}
	for _, next := range allowed[from] {
		if next == to {
			return true
		}
	}
	return false
}

// ChangeStatus applies a version-checked shipment status transition.
func (s *Service) ChangeStatus(ctx context.Context, id, status string, version int64) (models.Shipment, error) {
	v, err := s.Shipments.Get(ctx, id)
	if err != nil {
		return v, err
	}
	if v.Version != version || !CanTransition(v.Status, status) {
		return v, http.Fail(409, "Invalid shipment transition or stale version")
	}
	v.Status = status
	v.TrackingEvents = append(v.TrackingEvents, models.TrackingEvent{Status: status, OccurredAt: time.Now().UTC()})
	return s.Shipments.Replace(ctx, id, v)
}
