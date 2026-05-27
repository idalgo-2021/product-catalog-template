package product

import (
	"time"

	"github.com/shopspring/decimal"
)

type ImageResponse struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type ProductResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       decimal.Decimal   `json:"price"`
	Currency    string            `json:"currency"`
	Available   bool              `json:"available"`
	Thumbnail   ImageResponse     `json:"thumbnail"`
	Images      []ImageResponse   `json:"images"`
	Attributes  map[string]string `json:"attributes"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
