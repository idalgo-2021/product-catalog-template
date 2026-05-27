package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
)

type OrderService struct {
	orderRepo   OrderRepository
	productRepo ProductRepository
	logger      *zap.Logger
}

func NewOrderService(orderRepo OrderRepository, productRepo ProductRepository, logger *zap.Logger) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		logger:      logger,
	}
}

func (s *OrderService) GetOrder(ctx context.Context, id, userID string) (*domain.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !order.BelongsTo(userID) {
		return nil, fmt.Errorf("order %s: %w", id, domain.ErrForbidden)
	}

	return order, nil
}

func (s *OrderService) ListOrders(ctx context.Context, userID string, limit, offset int) ([]domain.Order, int, error) {
	orders, total, err := s.orderRepo.GetByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, total, nil
}

func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (*domain.Order, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: order must contain at least one item", domain.ErrInvalidInput)
	}

	// Validate all items first.
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return nil, err
		}
	}

	items, total, currency, err := s.processOrderItems(ctx, items)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	order := &domain.Order{
		ID:        uuid.New().String(),
		UserID:    userID,
		Items:     items,
		Status:    domain.StatusPending,
		Total:     total,
		Currency:  currency,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	s.logger.Info("order created", zap.String("order_id", order.ID), zap.String("user_id", userID))
	return order, nil
}

func (s *OrderService) UpdateOrder(ctx context.Context, id, userID string, input domain.UpdateOrderInput) (*domain.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !order.BelongsTo(userID) {
		return nil, fmt.Errorf("order %s: %w", id, domain.ErrForbidden)
	}

	// Update status if provided.
	if input.Status != nil {
		newStatus := *input.Status
		if !newStatus.IsValid() {
			return nil, fmt.Errorf("%w: invalid status: %s", domain.ErrInvalidInput, newStatus)
		}
		if !order.CanTransitionTo(newStatus) {
			return nil, fmt.Errorf("%w: cannot transition from %s to %s", domain.ErrInvalidInput, order.Status, newStatus)
		}
		order.Status = newStatus
	}

	// Update items if provided.
	if input.Items != nil {
		items := *input.Items

		if len(items) == 0 {
			return nil, fmt.Errorf("%w: order must contain at least one item", domain.ErrInvalidInput)
		}

		// Validate all items first.
		for _, item := range items {
			if err := item.Validate(); err != nil {
				return nil, err
			}
		}

		var total decimal.Decimal
		var currency string
		items, total, currency, err = s.processOrderItems(ctx, items)
		if err != nil {
			return nil, err
		}

		order.Items = items
		order.Total = total
		order.Currency = currency
	}

	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to update order: %w", err)
	}

	s.logger.Info("order updated", zap.String("order_id", order.ID))
	return order, nil
}

func (s *OrderService) processOrderItems(ctx context.Context, items []domain.OrderItem) ([]domain.OrderItem, decimal.Decimal, string, error) {
	// Batch-load all products in a single query.
	productIDs := make([]string, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	products, err := s.productRepo.GetByIDs(ctx, productIDs)
	if err != nil {
		return nil, decimal.Decimal{}, "", fmt.Errorf("failed to load products: %w", err)
	}

	var total decimal.Decimal
	var currency string

	for i, item := range items {
		product, ok := products[item.ProductID]
		if !ok {
			return nil, decimal.Decimal{}, "", fmt.Errorf("product %s: %w", item.ProductID, domain.ErrNotFound)
		}

		if !product.Available {
			return nil, decimal.Decimal{}, "", fmt.Errorf("%w: product not available: %s", domain.ErrInvalidInput, product.Name)
		}

		items[i].Price = product.Price
		total = total.Add(product.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
		currency = product.Currency
	}

	return items, total, currency, nil
}
