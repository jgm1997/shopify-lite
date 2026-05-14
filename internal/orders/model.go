package orders

import (
	"context"
	"database/sql"
	"errors"
	"shopify-lite/internal/db"
	"shopify-lite/internal/notifications"
	"time"
)

type OrderStatus string

const (
	Pending   OrderStatus = "pending"
	Confirmed OrderStatus = "confirmed"
	Shipped   OrderStatus = "shipped"
	Delivered OrderStatus = "delivered"
	Cancelled OrderStatus = "cancelled"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrEmptyItems        = errors.New("order must contain at least one item")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrInvalidStatus     = errors.New("invalid order status")
)

var validTransitions = map[OrderStatus][]OrderStatus{
	Pending:   {Confirmed, Cancelled},
	Confirmed: {Shipped, Cancelled},
	Shipped:   {Delivered},
	Delivered: {},
	Cancelled: {},
}

func (s OrderStatus) canTransitionTo(next OrderStatus) bool {
	for _, allowed := range validTransitions[s] {
		if allowed == next {
			return true
		}
	}
	return false
}

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

type BulkOrderItem struct {
	ProductID int32 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

type BulkOrderRequest struct {
	Items []BulkOrderItem `json:"items"`
}

type BulkOrderResult struct {
	ProductID int32  `json:"productId"`
	Status    string `json:"status"`
	Reason    string `json:"reason,omitempty"`
}

type BulkOrderResponse struct {
	Results   []BulkOrderResult `json:"results"`
	Succeeded int               `json:"succeeded"`
	Failed    int               `json:"failed"`
}

type Orders interface {
	PlaceOrder(ctx context.Context, customerID int, req PlaceOrderRequest) (Order, bool, error)
	GetCustomerOrders(ctx context.Context, customerID int) ([]Order, error)
	GetOrderByID(ctx context.Context, orderID, customerID int) (Order, bool, error)
	UpdateOrderStatus(ctx context.Context, orderID int, status string, merchantID int) (Order, error)
	ProcessBulkOrder(ctx context.Context, customerID int, req BulkOrderRequest) (BulkOrderResponse, error)
}

// Test interface
type orderPlace interface {
	PlaceOrder(ctx context.Context, customerID int, req PlaceOrderRequest) (Order, bool, error)
}

type Store struct {
	queries *db.Queries
	db      *sql.DB
}

type Handler struct {
	orders   Orders
	notifier *notifications.Notifier
}

func NewPsqlHandler(db *sql.DB, queries *db.Queries) *Store {
	return &Store{db: db, queries: queries}
}

func NewHandler(orders Orders, notifier *notifications.Notifier) *Handler {
	return &Handler{orders: orders, notifier: notifier}
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
