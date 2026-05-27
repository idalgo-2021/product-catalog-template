package service

import (
	"context"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
)

// UserRepository defines the contract for user persistence operations.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
}

// ProductRepository defines the contract for product persistence operations.
type ProductRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetByIDs(ctx context.Context, ids []string) (map[string]*domain.Product, error)
	List(ctx context.Context, limit, offset int) ([]domain.Product, int, error)
}

// OrderRepository defines the contract for order persistence operations.
type OrderRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	Create(ctx context.Context, order *domain.Order) error
	Update(ctx context.Context, order *domain.Order) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Order, int, error)
}
