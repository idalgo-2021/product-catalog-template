package service

import (
	"context"

	"go.uber.org/zap"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
)

type ProductService struct {
	repo   ProductRepository
	logger *zap.Logger
}

func NewProductService(repo ProductRepository, logger *zap.Logger) *ProductService {
	return &ProductService{
		repo:   repo,
		logger: logger,
	}
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *ProductService) ListProducts(ctx context.Context, limit, offset int) ([]domain.Product, int, error) {
	products, total, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return products, total, nil
}
