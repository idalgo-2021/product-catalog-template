package order

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderItemResponse struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     decimal.Decimal `json:"price"`
}

type OrderResponse struct {
	ID        string              `json:"id"`
	Status    string              `json:"status"`
	Total     decimal.Decimal     `json:"total"`
	Currency  string              `json:"currency"`
	Items     []OrderItemResponse `json:"items"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// отдельный DTO
type CreateOrderResponse struct {
	ID        string              `json:"id"`
	Status    string              `json:"status"`
	Total     decimal.Decimal     `json:"total"`
	Currency  string              `json:"currency"`
	Items     []OrderItemResponse `json:"items"`
	CreatedAt time.Time           `json:"created_at"`
}
