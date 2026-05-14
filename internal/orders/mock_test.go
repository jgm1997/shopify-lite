package orders_test

import (
	"context"
	"shopify-lite/internal/orders"
)

type mockOrders struct {
	placeOrderFn  func(ctx context.Context, customerID int, req orders.PlaceOrderRequest) (orders.Order, bool, error)
	processBulkFn func(ctx context.Context, customerID int, req orders.BulkOrderRequest) (orders.BulkOrderResponse, error)
}

func (m *mockOrders) PlaceOrder(ctx context.Context, customerID int, req orders.PlaceOrderRequest) (orders.Order, bool, error) {
	if m.placeOrderFn != nil {
		return m.placeOrderFn(ctx, customerID, req)
	}
	return orders.Order{}, false, nil
}

func (m *mockOrders) ProcessBulkOrder(ctx context.Context, customerID int, req orders.BulkOrderRequest) (orders.BulkOrderResponse, error) {
	if m.processBulkFn != nil {
		return m.processBulkFn(ctx, customerID, req)
	}
	return orders.BulkOrderResponse{}, nil
}

func (m *mockOrders) GetCustomerOrders(ctx context.Context, customerID int) ([]orders.Order, error) {
	return nil, nil
}

func (m *mockOrders) GetOrderByID(ctx context.Context, orderID, customerID int) (orders.Order, bool, error) {
	return orders.Order{}, false, nil
}

func (m *mockOrders) UpdateOrderStatus(ctx context.Context, orderID int, status string, merchantID int) (orders.Order, error) {
	return orders.Order{}, nil
}
