package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/idalgo-2021/product-catalog-backend/internal/auth"
	"github.com/idalgo-2021/product-catalog-backend/internal/domain"
)

type AuthService struct {
	userRepo           UserRepository
	jwtSecret          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	logger             *zap.Logger
}

func NewAuthService(
	userRepo UserRepository,
	jwtSecret string,
	accessTokenExpiry time.Duration,
	refreshTokenExpiry time.Duration,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		userRepo:           userRepo,
		jwtSecret:          jwtSecret,
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
		logger:             logger,
	}
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash from password: %w", err)
	}
	return string(hashedPassword), nil
}

func (s *AuthService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Info("user registered", zap.String("user_id", user.ID), zap.String("email", user.Email))
	return user, nil
}

// TODO: Зарефакторить
type TokenConfig struct {
	User               *domain.User
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", fmt.Errorf("%w: invalid credentials", domain.ErrUnauthorized)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", "", fmt.Errorf("%w: invalid credentials", domain.ErrUnauthorized)
	}

	accessToken, refreshToken, err := s.GenerateTokens(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate tokens for user %s: %w", user.ID, err)
	}

	s.logger.Info("user logged in", zap.String("user_id", user.ID))
	return accessToken, refreshToken, nil
}

func (s *AuthService) GenerateTokens(user *domain.User) (accessToken, refreshToken string, err error) {
	tc := TokenConfig{
		User:               user,
		Secret:             s.jwtSecret,
		AccessTokenExpiry:  s.accessTokenExpiry,
		RefreshTokenExpiry: s.refreshTokenExpiry,
	}

	return s.generateTokens(tc)
}

func (s *AuthService) generateTokens(tc TokenConfig) (accessToken, refreshToken string, err error) {
	accessToken, err = auth.GenerateToken(tc.User, tc.Secret, tc.AccessTokenExpiry, auth.TokenTypeAccess)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = auth.GenerateToken(tc.User, tc.Secret, tc.RefreshTokenExpiry, auth.TokenTypeRefresh)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := auth.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.TokenType != auth.TokenTypeRefresh {
		return "", "", fmt.Errorf("invalid token type: expected refresh token")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("user not found: %w", err)
	}

	newAccessToken, newRefreshToken, err := s.GenerateTokens(user)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate tokens for user %s: %w", user.ID, err)
	}

	s.logger.Info("tokens refreshed", zap.String("user_id", user.ID))
	return newAccessToken, newRefreshToken, nil
}

func (s *AuthService) ValidateToken(tokenString string) (string, error) {
	claims, err := auth.ValidateToken(tokenString, s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	return claims.UserID, nil
}
