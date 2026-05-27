package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Product struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       decimal.Decimal   `json:"price"`
	Currency    string            `json:"currency"`
	Available   bool              `json:"available"`
	Thumbnail   Image             `json:"thumbnail"`
	Images      []Image           `json:"images"`
	Attributes  map[string]string `json:"attributes"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
