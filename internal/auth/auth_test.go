package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret = "test-secret-key"
	correctPwd = "correctPassword123!"
)

func assertTokenClaims(t *testing.T, token string, expectedID int, expectedRole, secret string) {
	t.Helper()
	claims := &Claims{}
	parsedToken, _ := jwt.ParseWithClaims(token, claims, func(tok *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if parsedToken == nil || !parsedToken.Valid {
		return
	}
	if claims.UserID != expectedID {
		t.Errorf("Token UserID mismatch: got %d, want %d", claims.UserID, expectedID)
	}
	if claims.Role != expectedRole {
		t.Errorf("Token Role mismatch: got %s, want %s", claims.Role, expectedRole)
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "valid password",
			password:    "mySecurePassword123!",
			expectError: false,
		},
		{
			name:        "empty password",
			password:    "",
			expectError: false, // bcrypt can hash empty strings
		},
		{
			name:        "long password",
			password:    "this is a very long password that should still work fine",
			expectError: false,
		},
		{
			name:        "password with special chars",
			password:    "p@ssw0rd!#$%^&*()",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if (err != nil) != tt.expectError {
				t.Errorf("HashPassword() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if err == nil && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}

			// Hash should not equal the password
			if hash == tt.password {
				t.Error("HashPassword() returned plain text password")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		isValid  bool
	}{
		{
			name:     "correct password",
			password: correctPwd,
			isValid:  true,
		},
		{
			name:     "wrong password",
			password: "wrongPassword456!",
			isValid:  false,
		},
		{
			name:     "empty password check against non-empty hash",
			password: "somePassword",
			isValid:  false,
		},
		{
			name:     "very long password check",
			password: correctPwd,
			isValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(correctPwd)
			if err != nil {
				t.Fatalf("HashPassword failed: %v", err)
			}

			result := CheckPassword(hash, tt.password)

			if result != tt.isValid {
				t.Errorf("CheckPassword() got %v, want %v", result, tt.isValid)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		role      string
		secret    string
		expectErr bool
	}{
		{
			name:      "valid token generation",
			userID:    1,
			role:      "merchant",
			secret:    testSecret,
			expectErr: false,
		},
		{
			name:      "customer role",
			userID:    42,
			role:      "customer",
			secret:    testSecret,
			expectErr: false,
		},
		{
			name:      "empty secret",
			userID:    1,
			role:      "merchant",
			secret:    "",
			expectErr: false,
		},
		{
			name:      "zero user id",
			userID:    0,
			role:      "customer",
			secret:    testSecret,
			expectErr: false,
		},
		{
			name:      "negative user id",
			userID:    -1,
			role:      "merchant",
			secret:    testSecret,
			expectErr: false,
		},
		{
			name:      "long role string",
			userID:    99,
			role:      "super_admin_merchant_customer",
			secret:    testSecret,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.role, tt.secret)

			if (err != nil) != tt.expectErr {
				t.Errorf("GenerateToken() error = %v, expectErr %v", err, tt.expectErr)
				return
			}

			if err == nil && token == "" {
				t.Error("GenerateToken() returned empty token")
			}

			if err == nil {
				assertTokenClaims(t, token, tt.userID, tt.role, tt.secret)
			}
		})
	}
}

func TestVerifyToken(t *testing.T) {
	// Generate a valid token
	validToken, _ := GenerateToken(1, "merchant", testSecret)

	tests := []struct {
		name        string
		tokenStr    string
		secret      string
		expectError bool
		expectNil   bool
	}{
		{
			name:        "valid token",
			tokenStr:    validToken,
			secret:      testSecret,
			expectError: false,
			expectNil:   false,
		},
		{
			name:        "invalid token format",
			tokenStr:    "not.a.token",
			secret:      testSecret,
			expectError: true,
			expectNil:   true,
		},
		{
			name:        "wrong secret",
			tokenStr:    validToken,
			secret:      "wrong-secret",
			expectError: true,
			expectNil:   true,
		},
		{
			name:        "malformed token",
			tokenStr:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature",
			secret:      testSecret,
			expectError: true,
			expectNil:   true,
		},
		{
			name:        "valid token with empty string extracted",
			tokenStr:    "",
			secret:      testSecret,
			expectError: true,
			expectNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := VerifyToken(tt.tokenStr, tt.secret)

			if (err != nil) != tt.expectError {
				t.Errorf("VerifyToken() error = %v, expectError %v", err, tt.expectError)
			}

			if (claims == nil) != tt.expectNil {
				t.Errorf("VerifyToken() claims = %v, expectNil %v", claims, tt.expectNil)
			}

			if claims != nil && claims.UserID == 0 && claims.Role == "" {
				t.Error("VerifyToken() returned empty claims")
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	secret := testSecret

	// Generate token with standard 24 hour expiration
	token, err := GenerateToken(1, "merchant", secret)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := VerifyToken(token, secret)
	if err != nil {
		t.Fatalf("VerifyToken failed: %v", err)
	}

	// Check expiration is set to approximately 24 hours from now
	expiresAt := claims.ExpiresAt.Time
	now := time.Now()
	diff := expiresAt.Sub(now)

	// Should be between 23.9 and 24.1 hours
	minDiff := 23*time.Hour + 54*time.Minute
	maxDiff := 24*time.Hour + 6*time.Minute

	if diff < minDiff || diff > maxDiff {
		t.Errorf("Token expiration mismatch: got %v, want ~24h from now", diff)
	}
}

func TestHashPasswordAndCheckPasswordConsistency(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!#$%^&*()",
		},
		{
			name:     "long password within bcrypt limit",
			password: "this is a long password with special symbols that is under 72 bytes!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if err != nil {
				t.Fatalf("HashPassword failed: %v", err)
			}

			if !CheckPassword(hash, tt.password) {
				t.Error("CheckPassword failed with correct password")
			}

			if CheckPassword(hash, tt.password+"extra") {
				t.Error("CheckPassword succeeded with wrong password")
			}
		})
	}
}

func TestGenerateAndVerifyTokenConsistency(t *testing.T) {
	secret := testSecret

	tests := []struct {
		name   string
		userID int
		role   string
	}{
		{
			name:   "merchant user",
			userID: 1,
			role:   "merchant",
		},
		{
			name:   "customer user",
			userID: 2,
			role:   "customer",
		},
		{
			name:   "large user id",
			userID: 999999,
			role:   "merchant",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.role, secret)
			if err != nil {
				t.Fatalf("GenerateToken failed: %v", err)
			}

			claims, err := VerifyToken(token, secret)
			if err != nil {
				t.Fatalf("VerifyToken failed: %v", err)
			}

			if claims.UserID != tt.userID {
				t.Errorf("UserID mismatch: got %d, want %d", claims.UserID, tt.userID)
			}

			if claims.Role != tt.role {
				t.Errorf("Role mismatch: got %s, want %s", claims.Role, tt.role)
			}
		})
	}
}
