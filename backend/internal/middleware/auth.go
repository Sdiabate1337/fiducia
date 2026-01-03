package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/fiducia/backend/internal/config"
	"github.com/fiducia/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthContextKey type
type AuthContextKey string

const (
	// UserIDKey context key
	UserIDKey AuthContextKey = "user_id"
	// CabinetIDKey context key
	CabinetIDKey AuthContextKey = "cabinet_id"
	// RoleKey context key
	RoleKey AuthContextKey = "role"
)

// Auth middleware validates JWT tokens
func Auth(cfg *config.Config) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "authorization header required"}`, http.StatusUnauthorized)
				return
			}

			// Expect "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error": "invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			// Parse and validate token
			claims := &models.AuthClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(cfg.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				slog.Warn("invalid token", "error", err)
				http.Error(w, `{"error": "invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, CabinetIDKey, claims.CabinetID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID helper
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// GetCabinetID helper
func GetCabinetID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(CabinetIDKey).(uuid.UUID)
	return id, ok
}
