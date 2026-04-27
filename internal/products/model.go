package products

import (
	"context"
	"shopify-lite/internal/db"
	"time"
)

type Product struct {
	ID          int       `json:"id"`
	MerchantID  int       `json:"merchantId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ProductMetrics struct {
	MerchantID     int     `json:"merchantId"`
	TotalProducts  int     `json:"totalProducts"`
	TotalStock     int     `json:"totalStock"`
	InventoryValue float64 `json:"inventoryValue"`
	OutOfStock     int     `json:"outOfStock"`
}

type Products interface {
	GetMyProducts(ctx context.Context, merchantID int) ([]Product, error)
	CreateMyProduct(ctx context.Context, product Product) (Product, error)
	UpdateMyProduct(ctx context.Context, product Product) (Product, bool, error)
	DeleteMyProduct(ctx context.Context, productID, merchantID int) (bool, error)
	GetMyProductsDashboard(ctx context.Context, merchantID int) (ProductMetrics, error)
	GetProductsPaginated(ctx context.Context, page, limit int) ([]Product, error)
	GetProductByID(ctx context.Context, productID int) (Product, bool, error)
}

type Store struct{ queries *db.Queries }

type Handler struct {
	products Products
}

func NewPsqlHandler(queries *db.Queries) *Store {
	return &Store{queries: queries}
}

func toProduct(row db.Product) Product {
	return Product{
		ID:          int(row.ID),
		MerchantID:  int(row.MerchantID),
		Name:        row.Name,
		Description: row.Description.String,
		Price:       float64(row.Price),
		Stock:       int(row.Stock),
		CreatedAt:   row.CreatedAt,
	}
}

func NewHandler(products Products) *Handler {
	return &Handler{products: products}
}
