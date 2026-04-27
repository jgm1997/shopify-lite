package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"shopify-lite/internal/db"
)

const productErrFmt = "product %d: %w"

func mapProductErr(err error, productID int32) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf(productErrFmt, productID, ErrProductNotFound)
	}
	return fmt.Errorf(productErrFmt, productID, err)
}

func (psql *Store) PlaceOrder(ctx context.Context, customerID int, req PlaceOrderRequest) (Order, bool, error) {
	tx, err := psql.db.BeginTx(ctx, nil)
	if err != nil {
		return Order{}, false, err
	}
	defer tx.Rollback()
	qtx := psql.queries.WithTx(tx)

	var total float64
	type lockedProduct struct {
		row      db.GetProductForUpdateRow
		quantity int32
	}
	locked := make([]lockedProduct, 0, len(req.Items))

	for _, item := range req.Items {
		product, err := qtx.GetProductForUpdate(ctx, item.ProductID)
		if err != nil {
			return Order{}, false, mapProductErr(err, item.ProductID)
		}
		if product.Stock < item.Quantity {
			return Order{}, false, fmt.Errorf(productErrFmt, item.ProductID, ErrInsufficientStock)
		}
		total += float64(item.Quantity) * product.Price
		locked = append(locked, lockedProduct{row: product, quantity: item.Quantity})
	}

	order, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
		CustomerID: int32(customerID),
		Total:      total, // Will be updated after processing items
	})
	if err != nil {
		return Order{}, false, err
	}
	for _, lp := range locked {
		_, err = qtx.DeductStock(ctx, db.DeductStockParams{
			ID:    lp.row.ID,
			Stock: lp.quantity,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Order{}, false, mapProductErr(ErrInsufficientStock, lp.row.ID)
			}
		}

		// Create order items
		_, err = qtx.CreateOrderItem(ctx, db.CreateOrderItemParams{
			OrderID:   order.ID,
			ProductID: lp.row.ID,
			Quantity:  lp.quantity,
			UnitPrice: lp.row.Price,
		})
		if err != nil {
			return Order{}, false, err
		}
	}

	return toOrder(order), true, tx.Commit()
}

func (psql *Store) GetCustomerOrders(ctx context.Context, customerID int) ([]Order, error) {
	rows, err := psql.queries.GetCustomerOrders(ctx, int32(customerID))
	if err != nil {
		if err == sql.ErrNoRows {
			return []Order{}, nil
		}
		return []Order{}, err
	}
	orders := make([]Order, len(rows))
	for i, row := range rows {
		orders[i] = toOrder(db.Order{
			ID:         row.ID,
			CustomerID: row.CustomerID,
			Status:     db.OrderStatus(row.Status),
			Total:      row.Total,
			CreatedAt:  row.CreatedAt,
		})
	}
	return orders, nil
}

func (psql *Store) GetOrderByID(ctx context.Context, orderID, customerID int) (Order, bool, error) {
	row, err := psql.queries.GetOrderByID(ctx, db.GetOrderByIDParams{
		ID:         int32(orderID),
		CustomerID: int32(customerID),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return Order{}, false, nil
		}
		return Order{}, false, err
	}
	return toOrder(row), true, nil
}
