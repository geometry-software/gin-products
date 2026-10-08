package models

import "time"

// Entity contains persistence metadata shared by domain records.
type Entity struct {
	ID        string    `json:"id" bson:"_id"`
	Version   int64     `json:"version" bson:"version"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

// Product represents a catalog item and its stock.
type Product struct {
	Entity      `bson:",inline"`
	Name        string `json:"name" bson:"name" binding:"required,max=200"`
	Description string `json:"description" bson:"description" binding:"max=4000"`
	Price       int64  `json:"price" bson:"price" binding:"gte=0,lte=1000000000"`
	Quantity    int64  `json:"quantity" bson:"quantity" binding:"gte=0,lte=1000000000"`
	Active      bool   `json:"active" bson:"active"`
}

// Quantity identifies a product and requested amount.
type Quantity struct {
	ProductID string `json:"productId" bson:"productId" binding:"required,uuid"`
	Quantity  int64  `json:"quantity" bson:"quantity" binding:"gte=1,lte=1000000"`
}

// StockRequest describes an idempotent stock operation.
type StockRequest struct {
	OperationID string     `json:"operationId" binding:"required,uuid"`
	Items       []Quantity `json:"items" binding:"required,min=1,max=100,dive"`
}

// InvoiceItem stores an immutable product snapshot in an invoice.
type InvoiceItem struct {
	Quantity    `bson:",inline"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
	UnitPrice   int64  `json:"unitPrice" bson:"unitPrice"`
}

// Invoice represents an invoice and its lifecycle state.
type Invoice struct {
	Entity      `bson:",inline"`
	Name        string        `json:"name" bson:"name"`
	Description string        `json:"description" bson:"description"`
	Status      string        `json:"status" bson:"status"`
	Items       []InvoiceItem `json:"items" bson:"items"`
	Total       int64         `json:"total" bson:"total"`
}

// Recipient contains the shipment destination.
type Recipient struct {
	Name    string `json:"name" bson:"name" binding:"required,max=200"`
	Address string `json:"address" bson:"address" binding:"required,max=500"`
	City    string `json:"city" bson:"city" binding:"required,max=200"`
	Country string `json:"country" bson:"country" binding:"required,max=100"`
}

// TrackingEvent records a shipment status change.
type TrackingEvent struct {
	Status     string    `json:"status" bson:"status"`
	OccurredAt time.Time `json:"occurredAt" bson:"occurredAt"`
}

// Shipment represents a delivery and its tracking history.
type Shipment struct {
	Entity         `bson:",inline"`
	TrackingNumber string          `json:"trackingNumber" bson:"trackingNumber"`
	Status         string          `json:"status" bson:"status"`
	Recipient      Recipient       `json:"recipient" bson:"recipient"`
	Invoices       []Invoice       `json:"invoices" bson:"invoices"`
	CreatedBy      *User           `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	TrackingEvents []TrackingEvent `json:"trackingEvents" bson:"trackingEvents"`
}

// User represents a Firebase-backed account and its application role.
type User struct {
	Entity      `bson:",inline"`
	FirebaseUID string `json:"firebaseUid" bson:"firebaseUid"`
	Name        string `json:"name" bson:"name"`
	Email       string `json:"email" bson:"email"`
	Role        string `json:"role" bson:"role"`
	Active      bool   `json:"active" bson:"active"`
}

// Session represents a stored login session.
type Session struct {
	Entity    `bson:",inline"`
	UserID    string    `json:"userId" bson:"userId"`
	ExpiresAt time.Time `json:"expiresAt" bson:"expiresAt"`
}

// Page contains a paginated result set.
type Page[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
}

// Meta returns the embedded entity metadata.
func (e *Entity) Meta() *Entity { return e }
