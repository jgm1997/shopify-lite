package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"shopify-lite/internal/auth"
)

const (
	testSecret   = "test-secret"
	bearerPrefix = "Bearer "
)

func assertClaimsFromToken(t *testing.T, token string, expectedID int, expectedRole string) {
	t.Helper()
	m := NewAuthMiddleware(testSecret)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r)
		if claims == nil {
			t.Error("Expected claims, got nil")
			return
		}
		if claims.UserID != expectedID {
			t.Errorf("Expected UserID %d, got %d", expectedID, claims.UserID)
		}
		if claims.Role != expectedRole {
			t.Errorf("Expected Role %s, got %s", expectedRole, claims.Role)
		}
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", bearerPrefix+token)
	w := httptest.NewRecorder()
	m.Protect(nextHandler).ServeHTTP(w, req)
}

func TestProtectMiddleware(t *testing.T) {
	// Generate valid tokens
	validToken, _ := auth.GenerateToken(1, "customer", testSecret)
	merchantToken, _ := auth.GenerateToken(2, "merchant", testSecret)

	tests := []struct {
		name           string
		authHeader     string
		secret         string
		expectedStatus int
		shouldCallNext bool
	}{
		{
			name:           "valid token",
			authHeader:     bearerPrefix + validToken,
			secret:         testSecret,
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			secret:         testSecret,
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "invalid bearer format - no space",
			authHeader:     "Bearer" + validToken,
			secret:         testSecret,
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "invalid token signature",
			authHeader:     bearerPrefix + validToken,
			secret:         "wrong-secret",
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "malformed token",
			authHeader:     bearerPrefix + "invalid.token",
			secret:         testSecret,
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "merchant token",
			authHeader:     bearerPrefix + merchantToken,
			secret:         testSecret,
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "Bearer without token",
			authHeader:     bearerPrefix,
			secret:         testSecret,
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "Authorization header with different prefix",
			authHeader:     "Token " + validToken,
			secret:         testSecret,
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "empty secret with valid token format",
			authHeader:     bearerPrefix + validToken,
			secret:         "",
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("next"))
			})

			middleware := NewAuthMiddleware(tt.secret)
			handler := middleware.Protect(nextHandler)

			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if nextCalled != tt.shouldCallNext {
				t.Errorf("Expected next handler called = %v, got %v", tt.shouldCallNext, nextCalled)
			}
		})
	}
}

func TestRequireRoleMiddleware(t *testing.T) {
	customerToken, _ := auth.GenerateToken(1, "customer", testSecret)
	merchantToken, _ := auth.GenerateToken(2, "merchant", testSecret)
	adminToken, _ := auth.GenerateToken(3, "admin", testSecret)

	tests := []struct {
		name           string
		requiredRole   string
		authHeader     string
		secret         string
		expectedStatus int
		shouldCallNext bool
	}{
		{
			name:           "customer with customer role",
			requiredRole:   "customer",
			authHeader:     bearerPrefix + customerToken,
			secret:         testSecret,
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "merchant with merchant role",
			requiredRole:   "merchant",
			authHeader:     bearerPrefix + merchantToken,
			secret:         testSecret,
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "customer token but merchant required",
			requiredRole:   "merchant",
			authHeader:     bearerPrefix + customerToken,
			secret:         testSecret,
			expectedStatus: http.StatusForbidden,
			shouldCallNext: false,
		},
		{
			name:           "merchant token but customer required",
			requiredRole:   "customer",
			authHeader:     bearerPrefix + merchantToken,
			secret:         testSecret,
			expectedStatus: http.StatusForbidden,
			shouldCallNext: false,
		},
		{
			name:           "missing authorization header",
			requiredRole:   "merchant",
			authHeader:     "",
			secret:         testSecret,
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "invalid token",
			requiredRole:   "merchant",
			authHeader:     bearerPrefix + "invalid.token",
			secret:         testSecret,
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "wrong secret",
			requiredRole:   "merchant",
			authHeader:     bearerPrefix + merchantToken,
			secret:         "wrong-secret",
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
		{
			name:           "admin token requires admin role",
			requiredRole:   "admin",
			authHeader:     bearerPrefix + adminToken,
			secret:         testSecret,
			expectedStatus: http.StatusOK,
			shouldCallNext: true,
		},
		{
			name:           "customer requires admin role",
			requiredRole:   "admin",
			authHeader:     bearerPrefix + customerToken,
			secret:         testSecret,
			expectedStatus: http.StatusForbidden,
			shouldCallNext: false,
		},
		{
			name:           "invalid bearer format",
			requiredRole:   "merchant",
			authHeader:     "InvalidBearer " + merchantToken,
			secret:         testSecret,
			expectedStatus: http.StatusUnauthorized,
			shouldCallNext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("next"))
			})

			middleware := NewAuthMiddleware(tt.secret)
			handler := middleware.RequireRole(tt.requiredRole)(nextHandler)

			req := httptest.NewRequest("GET", "/merchant-only", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if nextCalled != tt.shouldCallNext {
				t.Errorf("Expected next handler called = %v, got %v", tt.shouldCallNext, nextCalled)
			}
		})
	}
}

func TestGetClaims(t *testing.T) {
	validToken, _ := auth.GenerateToken(42, "merchant", testSecret)

	t.Run("valid claims in context", func(t *testing.T) {
		assertClaimsFromToken(t, validToken, 42, "merchant")
	})

	t.Run("no claims in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		if GetClaims(req) != nil {
			t.Error("Expected nil claims, got non-nil")
		}
	})
}

func TestProtectMiddlewareContextPropagation(t *testing.T) {
	validToken, _ := auth.GenerateToken(10, "customer", testSecret)

	tests := []struct {
		name        string
		token       string
		checkClaims func(t *testing.T, claims *auth.Claims)
	}{
		{
			name:  "context has correct claims",
			token: validToken,
			checkClaims: func(t *testing.T, claims *auth.Claims) {
				if claims == nil {
					t.Error("Expected claims in context, got nil")
					return
				}
				if claims.UserID != 10 {
					t.Errorf("Expected UserID 10, got %d", claims.UserID)
				}
				if claims.Role != "customer" {
					t.Errorf("Expected role customer, got %s", claims.Role)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims := GetClaims(r)
				tt.checkClaims(t, claims)
				w.WriteHeader(http.StatusOK)
			})

			middleware := NewAuthMiddleware(testSecret)
			handler := middleware.Protect(nextHandler)

			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", bearerPrefix+tt.token)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestRequireRoleContextPropagation(t *testing.T) {
	merchantToken, _ := auth.GenerateToken(20, "merchant", testSecret)

	tests := []struct {
		name        string
		role        string
		token       string
		checkClaims func(t *testing.T, claims *auth.Claims)
	}{
		{
			name:  "merchant context preserved",
			role:  "merchant",
			token: merchantToken,
			checkClaims: func(t *testing.T, claims *auth.Claims) {
				if claims == nil {
					t.Error("Expected claims in context, got nil")
					return
				}
				if claims.UserID != 20 {
					t.Errorf("Expected UserID 20, got %d", claims.UserID)
				}
				if claims.Role != "merchant" {
					t.Errorf("Expected role merchant, got %s", claims.Role)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims := GetClaims(r)
				tt.checkClaims(t, claims)
				w.WriteHeader(http.StatusOK)
			})

			middleware := NewAuthMiddleware(testSecret)
			handler := middleware.RequireRole(tt.role)(nextHandler)

			req := httptest.NewRequest("GET", "/merchant-only", nil)
			req.Header.Set("Authorization", bearerPrefix+tt.token)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestNewAuthMiddleware(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{
			name:   "with non-empty secret",
			secret: "my-secret-key",
		},
		{
			name:   "with empty secret",
			secret: "",
		},
		{
			name:   "with long secret",
			secret: "a" + string(make([]byte, 1000)) + "z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewAuthMiddleware(tt.secret)
			if middleware == nil {
				t.Error("NewAuthMiddleware returned nil")
			}
		})
	}
}
