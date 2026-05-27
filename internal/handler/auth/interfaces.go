package auth

import (
	"context"

	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (accessToken string, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (accessToken string, newRefreshToken string, err error)
	ValidateToken(tokenString string) (string, error)
}
