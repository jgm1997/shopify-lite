package merchants

import (
	"shopify-lite/internal/db"
	"time"
)

type Merchant struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	StoreName   string    `json:"store_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type Store struct{ queries *db.Queries }

type Handler struct {
	merchants Merchants
	jwtSecret string
}

func NewHandler(merchants Merchants, jwtSecret string) *Handler {
	return &Handler{merchants: merchants, jwtSecret: jwtSecret}
}
