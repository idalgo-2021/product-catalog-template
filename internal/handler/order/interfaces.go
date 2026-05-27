package order

import (
	"context"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
)

type OrderService interface {
	GetOrder(ctx context.Context, id, userID string) (*domain.Order, error)
	ListOrders(ctx context.Context, userID string, limit, offset int) ([]domain.Order, int, error)
	CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (*domain.Order, error)
	UpdateOrder(ctx context.Context, id, userID string, input domain.UpdateOrderInput) (*domain.Order, error)
}
