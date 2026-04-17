package users

import (
	"time"

	"shopify-lite/internal/db"
)

type UserRole string

const (
	UserRoleMerchant UserRole = "merchant"
	UserRoleCustomer UserRole = "customer"
)

type PSQL struct {
	queries *db.Queries
}

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type Handler struct {
	users     Users
	jwtSecret string
}

func NewHandler(users Users, jwtSecret string) *Handler {
	return &Handler{users: users, jwtSecret: jwtSecret}
}
