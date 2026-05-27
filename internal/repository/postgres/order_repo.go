package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
)

type OrderRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewOrderRepo(db *pgxpool.Pool, logger *zap.Logger) *OrderRepo {
	return &OrderRepo{
		db:     db,
		logger: logger,
	}
}

func (r *OrderRepo) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	query := `
		SELECT id, user_id, items, status, total, currency, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order domain.Order
	var itemsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.UserID,
		&itemsJSON,
		&order.Status,
		&order.Total,
		&order.Currency,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order %s: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal items: %w", err)
	}

	return &order, nil
}

func (r *OrderRepo) Create(ctx context.Context, order *domain.Order) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	query := `
		INSERT INTO orders (id, user_id, items, status, total, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = r.db.Exec(ctx, query,
		order.ID,
		order.UserID,
		itemsJSON,
		order.Status,
		order.Total,
		order.Currency,
		order.CreatedAt,
		order.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

func (r *OrderRepo) Update(ctx context.Context, order *domain.Order) error {
	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	query := `
		UPDATE orders
		SET items = $1, status = $2, total = $3, currency = $4, updated_at = $5
		WHERE id = $6
	`

	tag, err := r.db.Exec(ctx, query,
		itemsJSON,
		order.Status,
		order.Total,
		order.Currency,
		order.UpdatedAt,
		order.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("order %s: %w", order.ID, domain.ErrNotFound)
	}

	return nil
}

func (r *OrderRepo) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Order, int, error) {
	query := `
		SELECT id, user_id, items, status, total, currency, created_at, updated_at, COUNT(*) OVER()
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order
	var total int
	for rows.Next() {
		var order domain.Order
		var itemsJSON []byte

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&itemsJSON,
			&order.Status,
			&order.Total,
			&order.Currency,
			&order.CreatedAt,
			&order.UpdatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}

		if err := json.Unmarshal(itemsJSON, &order.Items); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal items: %w", err)
		}

		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate orders: %w", err)
	}

	return orders, total, nil
}
