package merchants

import (
	"context"
	"database/sql"
	"shopify-lite/internal/db"
)

type Merchants interface {
	GetMerchantByUserID(ctx context.Context, userID int) (Merchant, bool, error)
	CreateMerchant(ctx context.Context, merchant Merchant) (Merchant, error)
}

func NewPsqlHandler(queries *db.Queries) *Store {
	return &Store{queries: queries}
}

func toMerchant(row db.Merchant) Merchant {
	return Merchant{
		ID:          int(row.ID),
		UserID:      int(row.UserID),
		StoreName:   row.StoreName,
		Description: row.Description.String,
		CreatedAt:   row.CreatedAt,
	}
}

func (psql *Store) GetMerchantByUserID(ctx context.Context, userID int) (Merchant, bool, error) {
	row, err := psql.queries.GetMerchantByUserID(ctx, int32(userID))
	if err != nil {
		if err == sql.ErrNoRows {
			return Merchant{}, false, nil
		}
		return Merchant{}, false, err
	}
	return toMerchant(row), true, nil
}

func (psql *Store) CreateMerchant(ctx context.Context, merchant Merchant) (Merchant, error) {
	existing, found, err := psql.GetMerchantByUserID(ctx, merchant.UserID)
	if err != nil {
		return Merchant{}, err
	}
	if found {
		return existing, err
	}
	created, err := psql.queries.CreateMerchant(ctx, db.CreateMerchantParams{
		UserID:    int32(merchant.UserID),
		StoreName: merchant.StoreName,
		Description: sql.NullString{
			String: merchant.Description,
			Valid:  merchant.Description != "",
		},
	})
	if err != nil {
		return Merchant{}, err
	}
	return toMerchant(created), nil
}
