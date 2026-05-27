// internal/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/idalgo-2021/product-catalog-backend/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// GetUserID extracts the authenticated user ID from the request context.
// Returns empty string and false if not present (e.g. unauthenticated request).
func GetUserID(ctx context.Context) (string, bool) {
	uid, ok := ctx.Value(UserIDKey).(string)
	return uid, ok
}

func Auth(jwtSecret string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Authorization header required"}}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid authorization header format"}}`, http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateToken(parts[1], jwtSecret)
			if err != nil {
				logger.Warn("invalid token", zap.Error(err))
				http.Error(w, `{"error":{"code":"UNAUTHORIZED","message":"Invalid token"}}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
