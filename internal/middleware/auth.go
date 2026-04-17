package middleware

import (
	"context"
	"net/http"
	"strings"

	"shopify-lite/internal/auth"
)

type contextKey string

const ClaimsKey contextKey = "claims"

type AuthMiddleware struct {
	secret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

// Protect wraps a handler to require valid JWT token
func (m *AuthMiddleware) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == header {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		claims, err := auth.VerifyToken(tokenStr, m.secret)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole wraps a handler to require valid JWT token with specific role
func (m *AuthMiddleware) RequireRole(role string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, "authorization header required", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		if tokenStr == header {
			http.Error(w, "invalid authorization format", http.StatusUnauthorized)
			return
		}

		claims, err := auth.VerifyToken(tokenStr, m.secret)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		if claims.Role != role {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClaims extracts claims from request context
func GetClaims(r *http.Request) *auth.Claims {
	claims, ok := r.Context().Value(ClaimsKey).(*auth.Claims)
	if !ok {
		return nil
	}
	return claims
}
