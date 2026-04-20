package users

import (
	"shopify-lite/internal/db"
	"time"
)

type UserRole string

const (
	RoleMerchant UserRole = "merchant"
	RoleCustomer UserRole = "customer"
)

type User struct {
	ID        int
	Email     string
	Password  string
	Role      UserRole
	CreatedAt time.Time
}

type UserResponse struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

type Store struct{ queries *db.Queries }

type Handler struct {
	users     Users
	jwtSecret string
}

func NewHandler(users Users, jwtSecret string) *Handler {
	return &Handler{users: users, jwtSecret: jwtSecret}
}
