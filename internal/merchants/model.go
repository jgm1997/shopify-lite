package merchants

import (
	"shopify-lite/internal/db"
	"time"
)

type Merchant struct {
	ID          int       `json:"id"`
	UserID      int       `json:"userId"`
	StoreName   string    `json:"storeName"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Store struct{ queries *db.Queries }

type Handler struct {
	merchants Merchants
}

func NewHandler(merchants Merchants) *Handler {
	return &Handler{merchants: merchants}
}
