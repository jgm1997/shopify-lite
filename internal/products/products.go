package products

import (
	"context"
	"database/sql"
	"shopify-lite/internal/db"
	"shopify-lite/internal/utils"
)

func (psql *Store) GetMerchantIDByUserID(ctx context.Context, userID int) (int, error) {
	merchant, err := psql.queries.GetMerchantByUserID(ctx, int32(userID))
	if err != nil {
		return 0, err
	}
	return int(merchant.ID), nil
}

func (psql *Store) GetMyProducts(ctx context.Context, merchantID int) ([]Product, error) {
	rows, err := psql.queries.GetMerchantsProducts(ctx, int32(merchantID))
	if err != nil {
		return []Product{}, err
	}
	products := make([]Product, len(rows))
	for i, row := range rows {
		products[i] = toProduct(row)
	}
	return products, nil
}

func (psql *Store) CreateMyProduct(ctx context.Context, product Product) (Product, error) {
	created, err := psql.queries.CreateProduct(ctx, db.CreateProductParams{
		MerchantID: int32(product.MerchantID),
		Name:       product.Name,
		Description: sql.NullString{
			String: product.Description,
			Valid:  product.Description != "",
		},
		Price: float64(product.Price),
		Stock: int32(product.Stock),
	})
	if err != nil {
		return Product{}, err
	}
	return toProduct(created), nil
}

func (psql *Store) UpdateMyProduct(ctx context.Context, product Product) (Product, bool, error) {
	row, err := psql.queries.UpdateProduct(ctx, db.UpdateProductParams{
		ID:         int32(product.ID),
		MerchantID: int32(product.MerchantID),
		Name:       product.Name,
		Description: sql.NullString{
			String: product.Description,
			Valid:  product.Description != "",
		},
		Price: float64(product.Price),
		Stock: int32(product.Stock),
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return Product{}, false, nil
		}
		return Product{}, false, err
	}
	return toProduct(row), true, nil
}

func (psql *Store) DeleteMyProduct(ctx context.Context, productID, merchantID int) (bool, error) {
	n, err := psql.queries.DeleteProduct(ctx, db.DeleteProductParams{
		ID:         int32(productID),
		MerchantID: int32(merchantID),
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (psql *Store) GetMyProductsDashboard(ctx context.Context, merchantID int) (ProductMetrics, error) {
	rows, err := psql.queries.GetMerchantDashboard(ctx, int32(merchantID))
	if err != nil {
		return ProductMetrics{}, err
	}
	metrics := ProductMetrics{
		MerchantID:     merchantID,
		TotalProducts:  int(rows.TotalProducts),
		TotalStock:     utils.ToInt(rows.TotalStock),
		InventoryValue: utils.ToFloat64(rows.InventoryValue),
		OutOfStock:     int(rows.OutOfStock),
	}
	return metrics, nil
}
