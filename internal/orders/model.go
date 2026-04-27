package orders

import (
	"context"
	"database/sql"
	"errors"
	"shopify-lite/internal/db"
	"time"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrEmptyItems        = errors.New("order must contain at least one item")
)

type OrderStatus string

const (
	Pending   OrderStatus = "pending"
	Shipped   OrderStatus = "shipped"
	Delivered OrderStatus = "delivered"
)

type Order struct {
	ID         int32       `json:"id"`
	CustomerID int32       `json:"customerId"`
	Status     OrderStatus `json:"status"`
	Total      float64     `json:"total"`
	CreatedAt  time.Time   `json:"createdAt"`
}

type PlaceOrderRequest struct {
	Items []OrderItemRequest `json:"items"`
}

type OrderItemRequest struct {
	ProductID int32 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

type OrderItem struct {
	ID        int32   `json:"id"`
	OrderID   int32   `json:"orderId"`
	ProductID int32   `json:"productId"`
	Quantity  int32   `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

type Orders interface {
	PlaceOrder(ctx context.Context, customerID int, req PlaceOrderRequest) (Order, bool, error)
	GetCustomerOrders(ctx context.Context, customerID int) ([]Order, error)
	GetOrderByID(ctx context.Context, orderID, customerID int) (Order, bool, error)
}

type Store struct {
	queries *db.Queries
	db      *sql.DB
}

type Handler struct{ orders Orders }

func NewPsqlHandler(db *sql.DB, queries *db.Queries) *Store {
	return &Store{db: db, queries: queries}
}

func NewHandler(orders Orders) *Handler {
	return &Handler{orders: orders}
}

func toOrder(row db.Order) Order {
	return Order{
		ID:         row.ID,
		CustomerID: row.CustomerID,
		Status:     OrderStatus(row.Status),
		Total:      row.Total,
		CreatedAt:  row.CreatedAt,
	}
}

func toOrderItem(row db.OrderItem) OrderItem {
	return OrderItem{
		ID:        row.ID,
		OrderID:   row.OrderID,
		ProductID: row.ProductID,
		Quantity:  row.Quantity,
		UnitPrice: row.UnitPrice,
	}
}
