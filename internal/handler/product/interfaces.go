package product

import (
	"context"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
)

type ProductService interface {
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	ListProducts(ctx context.Context, limit, offset int) ([]domain.Product, int, error)
}
