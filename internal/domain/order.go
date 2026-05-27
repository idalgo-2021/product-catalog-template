package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
}

// Validate checks that an order item has valid fields.
func (i OrderItem) Validate() error {
	if i.ProductID == "" {
		return fmt.Errorf("%w: product_id is required", ErrInvalidInput)
	}
	if i.Quantity <= 0 {
		return fmt.Errorf("%w: quantity must be positive", ErrInvalidInput)
	}
	return nil
}

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"
	StatusConfirmed OrderStatus = "confirmed"
	StatusShipped   OrderStatus = "shipped"
	StatusDelivered OrderStatus = "delivered"
	StatusCancelled OrderStatus = "cancelled"
)

// IsValid returns true if the status is a known OrderStatus value.
func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusShipped, StatusDelivered, StatusCancelled:
		return true
	}
	return false
}

// validTransitions defines the allowed state machine for order statuses.
var validTransitions = map[OrderStatus][]OrderStatus{
	StatusPending:   {StatusConfirmed, StatusCancelled},
	StatusConfirmed: {StatusShipped, StatusCancelled},
	StatusShipped:   {StatusDelivered},
}

// CanTransitionTo checks whether transitioning from the current status to newStatus is allowed.
func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	allowed, ok := validTransitions[o.Status]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == newStatus {
			return true
		}
	}
	return false
}

// CalculateTotal recomputes the order total from its items.
func (o *Order) CalculateTotal() {
	total := decimal.NewFromInt(0)
	for _, item := range o.Items {
		total = total.Add(item.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}
	o.Total = total
}

// BelongsTo checks whether the order belongs to the given user.
func (o *Order) BelongsTo(userID string) bool {
	return o.UserID == userID
}

// UpdateOrderInput is a typed struct for partial order updates.
// Both Status and Items are optional — only provided fields will be updated.
type UpdateOrderInput struct {
	Status *OrderStatus `json:"status,omitempty"`
	Items  *[]OrderItem `json:"items,omitempty"`
}

type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Items     []OrderItem `json:"items"`
	Status    OrderStatus `json:"status"`
	Total     decimal.Decimal `json:"total"`
	Currency  string      `json:"currency"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
