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

// scanProduct is a helper that scans a single product row and unmarshals JSONB fields.
func scanProduct(scannable interface {
	Scan(dest ...any) error
}, product *domain.Product) error {
	var thumbnailJSON, imagesJSON, attributesJSON []byte

	err := scannable.Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Currency,
		&product.Available,
		&thumbnailJSON,
		&imagesJSON,
		&attributesJSON,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(thumbnailJSON, &product.Thumbnail); err != nil {
		return fmt.Errorf("failed to unmarshal thumbnail: %w", err)
	}
	if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
		return fmt.Errorf("failed to unmarshal images: %w", err)
	}
	if err := json.Unmarshal(attributesJSON, &product.Attributes); err != nil {
		return fmt.Errorf("failed to unmarshal attributes: %w", err)
	}

	return nil
}

type ProductRepo struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewProductRepo(db *pgxpool.Pool, logger *zap.Logger) *ProductRepo {
	return &ProductRepo{
		db:     db,
		logger: logger,
	}
}

func (r *ProductRepo) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `
		SELECT id, name, description, price, currency, available, thumbnail, images, attributes, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product domain.Product
	if err := scanProduct(r.db.QueryRow(ctx, query, id), &product); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("product %s: %w", id, domain.ErrNotFound)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &product, nil
}

func (r *ProductRepo) GetByIDs(ctx context.Context, ids []string) (map[string]*domain.Product, error) {
	if len(ids) == 0 {
		return make(map[string]*domain.Product), nil
	}

	query := `
		SELECT id, name, description, price, currency, available, thumbnail, images, attributes, created_at, updated_at
		FROM products
		WHERE id = ANY($1)
	`

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by ids: %w", err)
	}
	defer rows.Close()

	result := make(map[string]*domain.Product, len(ids))
	for rows.Next() {
		var product domain.Product
		if err := scanProduct(rows, &product); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		result[product.ID] = &product
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate products: %w", err)
	}

	return result, nil
}

func (r *ProductRepo) List(ctx context.Context, limit, offset int) ([]domain.Product, int, error) {
	query := `
		SELECT id, name, description, price, currency, available, thumbnail, images, attributes, created_at, updated_at, COUNT(*) OVER()
		FROM products
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	var total int
	for rows.Next() {
		var product domain.Product
		var thumbnailJSON, imagesJSON, attributesJSON []byte

		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Currency,
			&product.Available,
			&thumbnailJSON,
			&imagesJSON,
			&attributesJSON,
			&product.CreatedAt,
			&product.UpdatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		if err := json.Unmarshal(thumbnailJSON, &product.Thumbnail); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal thumbnail: %w", err)
		}
		if err := json.Unmarshal(imagesJSON, &product.Images); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal images: %w", err)
		}
		if err := json.Unmarshal(attributesJSON, &product.Attributes); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal attributes: %w", err)
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate products: %w", err)
	}

	return products, total, nil
}
