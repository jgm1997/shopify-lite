package users

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"shopify-lite/internal/auth"
)

const (
	testJWTSecret        = "test-secret"
	testUserEmail        = "user@example.com"
	testMerchantEmail    = "merchant@example.com"
	testSecurePassword   = "SecurePassword123!"
	testCorrectPassword  = "correctPassword123!"
	bearerPrefix         = "Bearer "
	errUnmarshalFmt      = "Failed to unmarshal response: %v"
	errExpectedStatusFmt = "Expected status %d, got %d"
	errDatabaseErrorMsg  = "database error"
)

func assertTokenInBody(t *testing.T, body string) {
	t.Helper()
	var resp map[string]interface{}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf(errUnmarshalFmt, err)
	}
	if _, ok := resp["token"]; !ok {
		t.Error("Response missing token field")
	}
}

func assertStatus(t *testing.T, want, got int) {
	t.Helper()
	if want != got {
		t.Errorf(errExpectedStatusFmt, want, got)
	}
}

func mockCreateUserFn(wantErr error, resp User) func(context.Context, User) (User, error) {
	return func(_ context.Context, _ User) (User, error) {
		if wantErr != nil {
			return User{}, wantErr
		}
		return resp, nil
	}
}

// MockUsersStore implements Users interface for testing
type MockUsersStore struct {
	createUserFunc     func(ctx context.Context, u User) (User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (User, bool, error)
	getUserByIDFunc    func(ctx context.Context, id int) (User, bool, error)
}

func (m *MockUsersStore) CreateUser(ctx context.Context, u User) (User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, u)
	}
	return User{}, nil
}

func (m *MockUsersStore) GetUserByEmail(ctx context.Context, email string) (User, bool, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return User{}, false, nil
}

func (m *MockUsersStore) GetUserByID(ctx context.Context, id int) (User, bool, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return User{}, false, nil
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name               string
		email              string
		password           string
		role               string
		mockCreateUserErr  error
		mockCreateUserResp User
		expectedStatus     int
		checkResponse      func(t *testing.T, body string)
	}{
		{
			name:     "valid registration",
			email:    testUserEmail,
			password: testSecurePassword,
			role:     "customer",
			mockCreateUserResp: User{
				ID:        1,
				Email:     testUserEmail,
				Role:      "customer",
				CreatedAt: time.Now(),
			},
			expectedStatus: http.StatusCreated,
			checkResponse:  assertTokenInBody,
		},
		{
			name:           "invalid email - too short",
			email:          "a@b",
			password:       testSecurePassword,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid email - no @",
			email:          "userexample.com",
			password:       testSecurePassword,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid email - too long",
			email:          "a" + string(make([]byte, 250)) + "@example.com",
			password:       testSecurePassword,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid password - too short",
			email:          testUserEmail,
			password:       "short",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:              "database error on create",
			email:             testUserEmail,
			password:          testSecurePassword,
			mockCreateUserErr: ErrUserAlreadyExists,
			expectedStatus:    http.StatusInternalServerError,
		},
		{
			name:     "merchant role",
			email:    testMerchantEmail,
			password: "MerchantPass123!",
			role:     "merchant",
			mockCreateUserResp: User{
				ID:        2,
				Email:     testMerchantEmail,
				Role:      "merchant",
				CreatedAt: time.Now(),
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty request body",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.email != "" || tt.password != "" {
				user := User{
					Email:    tt.email,
					Password: tt.password,
					Role:     UserRole(tt.role),
				}
				reqBody, _ = json.Marshal(user)
			} else {
				reqBody = []byte("invalid json")
			}

			mock := &MockUsersStore{
				createUserFunc: mockCreateUserFn(tt.mockCreateUserErr, tt.mockCreateUserResp),
			}

			handler := NewHandler(mock, testJWTSecret)

			req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RegisterHandler(w, req)

			assertStatus(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	hashedPassword, _ := auth.HashPassword(testCorrectPassword)

	tests := []struct {
		name            string
		email           string
		password        string
		mockGetUserFunc func(ctx context.Context, email string) (User, bool, error)
		expectedStatus  int
		checkResponse   func(t *testing.T, body string)
	}{
		{
			name:     "successful login",
			email:    testUserEmail,
			password: testCorrectPassword,
			mockGetUserFunc: func(_ context.Context, _ string) (User, bool, error) {
				return User{
					ID:       1,
					Email:    testUserEmail,
					Password: hashedPassword,
					Role:     "customer",
				}, true, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse:  assertTokenInBody,
		},
		{
			name:     "wrong password",
			email:    testUserEmail,
			password: "wrongPassword123!",
			mockGetUserFunc: func(_ context.Context, _ string) (User, bool, error) {
				return User{
					ID:       1,
					Email:    testUserEmail,
					Password: hashedPassword,
					Role:     "customer",
				}, true, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:     "user not found",
			email:    "nonexistent@example.com",
			password: "somePassword123!",
			mockGetUserFunc: func(_ context.Context, _ string) (User, bool, error) {
				return User{}, false, nil
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty email",
			email:          "",
			password:       testCorrectPassword,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty password",
			email:          testUserEmail,
			password:       "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     errDatabaseErrorMsg,
			email:    testUserEmail,
			password: testCorrectPassword,
			mockGetUserFunc: func(_ context.Context, _ string) (User, bool, error) {
				return User{}, false, ErrDatabaseError
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:     "merchant role login",
			email:    testMerchantEmail,
			password: testCorrectPassword,
			mockGetUserFunc: func(_ context.Context, _ string) (User, bool, error) {
				return User{
					ID:       2,
					Email:    testMerchantEmail,
					Password: hashedPassword,
					Role:     "merchant",
				}, true, nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockUsersStore{
				getUserByEmailFunc: tt.mockGetUserFunc,
			}

			handler := NewHandler(mock, testJWTSecret)

			body := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}
			reqBody, _ := json.Marshal(body)

			req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.LoginHandler(w, req)

			assertStatus(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

func TestGetMeHandler(t *testing.T) {
	// Generate valid token
	validToken, _ := auth.GenerateToken(1, "customer", testJWTSecret)
	tokenWithMerchantRole, _ := auth.GenerateToken(2, "merchant", testJWTSecret)

	tests := []struct {
		name            string
		authHeader      string
		mockGetUserFunc func(ctx context.Context, id int) (User, bool, error)
		expectedStatus  int
		checkResponse   func(t *testing.T, body string)
	}{
		{
			name:       "successful get me",
			authHeader: bearerPrefix + validToken,
			mockGetUserFunc: func(_ context.Context, _ int) (User, bool, error) {
				return User{
					ID:        1,
					Email:     testUserEmail,
					Role:      "customer",
					CreatedAt: time.Now(),
				}, true, nil
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var user User
				if err := json.Unmarshal([]byte(body), &user); err != nil {
					t.Fatalf(errUnmarshalFmt, err)
				}
				if user.ID != 1 {
					t.Errorf("Expected user ID 1, got %d", user.ID)
				}
				if user.Email != testUserEmail {
					t.Errorf("Expected email %s, got %s", testUserEmail, user.Email)
				}
			},
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token format",
			authHeader:     "InvalidToken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed bearer token",
			authHeader:     bearerPrefix + "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "user not found",
			authHeader: bearerPrefix + validToken,
			mockGetUserFunc: func(_ context.Context, _ int) (User, bool, error) {
				return User{}, false, nil
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       errDatabaseErrorMsg,
			authHeader: bearerPrefix + validToken,
			mockGetUserFunc: func(_ context.Context, _ int) (User, bool, error) {
				return User{}, false, ErrDatabaseError
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "merchant user get me",
			authHeader: bearerPrefix + tokenWithMerchantRole,
			mockGetUserFunc: func(_ context.Context, _ int) (User, bool, error) {
				return User{
					ID:        2,
					Email:     testMerchantEmail,
					Role:      "merchant",
					CreatedAt: time.Now(),
				}, true, nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockUsersStore{
				getUserByIDFunc: tt.mockGetUserFunc,
			}

			handler := NewHandler(mock, testJWTSecret)

			req := httptest.NewRequest("GET", "/api/v1/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.GetMeHandler(w, req)

			assertStatus(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.String())
			}
		})
	}
}

// Test validation functions
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{
			name:  "valid email",
			email: "user@example.com",
			valid: true,
		},
		{
			name:  "valid email with subdomain",
			email: "user@mail.example.com",
			valid: true,
		},
		{
			name:  "email too short",
			email: "a@b",
			valid: false,
		},
		{
			name:  "email without @",
			email: "userexample.com",
			valid: false,
		},
		{
			name:  "email too long",
			email: "a" + string(make([]byte, 250)) + "@example.com",
			valid: false,
		},
		{
			name:  "empty email",
			email: "",
			valid: false,
		},
		{
			name:  "email with multiple @",
			email: "user@@example.com",
			valid: true, // contains @ so valid by our simple check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateEmail(tt.email)
			if result != tt.valid {
				t.Errorf("validateEmail(%s) = %v, want %v", tt.email, result, tt.valid)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{
			name:     "valid password",
			password: "SecurePassword123!",
			valid:    true,
		},
		{
			name:     "minimum length password",
			password: "12345678",
			valid:    true,
		},
		{
			name:     "password too short",
			password: "short7",
			valid:    false,
		},
		{
			name:     "empty password",
			password: "",
			valid:    false,
		},
		{
			name:     "long password",
			password: "ThisIsAVeryLongPasswordWithManyCharactersAndSymbols!@#$%^&*()",
			valid:    true,
		},
		{
			name:     "exactly 7 chars",
			password: "1234567",
			valid:    false,
		},
		{
			name:     "exactly 8 chars",
			password: "12345678",
			valid:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validatePassword(tt.password)
			if result != tt.valid {
				t.Errorf("validatePassword(%s) = %v, want %v", tt.password, result, tt.valid)
			}
		})
	}
}

var (
	ErrUserAlreadyExists = testError("user already exists")
	ErrDatabaseError     = testError(errDatabaseErrorMsg)
)

type testError string

func (e testError) Error() string {
	return string(e)
}
