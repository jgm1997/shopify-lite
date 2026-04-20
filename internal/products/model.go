package products

import (
	"context"
	"shopify-lite/internal/db"
)

type Product struct {
	ID          int     `json:"id"`
	MerchantID  int     `json:"merchant_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type ProductMetrics struct {
	MerchantID     int     `json:"merchant_id"`
	TotalProducts  int     `json:"total_products"`
	TotalStock     int     `json:"total_stock"`
	InventoryValue float64 `json:"inventory_value"`
	OutOfStock     int     `json:"out_of_stock"`
}

type Products interface {
	GetMerchantIDByUserID(ctx context.Context, userID int) (int, error)
	GetMyProducts(ctx context.Context, merchantID int) ([]Product, error)
	CreateMyProduct(ctx context.Context, product Product) (Product, error)
	UpdateMyProduct(ctx context.Context, product Product) (Product, error)
	DeleteMyProduct(ctx context.Context, productID int) (bool, error)
	GetMyProductsDashboard(ctx context.Context, merchantID int) (ProductMetrics, error)
}

type Store struct{ queries *db.Queries }

type Handler struct {
	products  Products
	jwtSecret string
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
	}
}

func NewHandler(products Products, jwtSecret string) *Handler {
	return &Handler{products: products, jwtSecret: jwtSecret}
}
