package users

import (
	"context"
	"database/sql"
	"strings"
	"shopify-lite/internal/auth"
	"shopify-lite/internal/db"
)

type Users interface {
	CreateUser(ctx context.Context, u User) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, bool, error)
	GetUserByID(ctx context.Context, id int) (User, bool, error)
}

func NewPsqlHandler(queries *db.Queries) *Store {
	return &Store{
		queries: queries,
	}
}

func toUser(row db.User) User {
	return User{
		ID:        int(row.ID),
		Email:     row.Email,
		Password:  row.PasswordHash,
		Role:      UserRole(row.Role),
		CreatedAt: row.CreatedAt,
	}
}

func (psql *Store) CreateUser(ctx context.Context, u User) (User, error) {
	passwordHash, err := auth.HashPassword(u.Password)
	if err != nil {
		return User{}, err
	}
	created, err := psql.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        u.Email,
		PasswordHash: passwordHash,
		Role:         db.UserRole(u.Role),
	})
	if err != nil {
		return User{}, err
	}
	if u.Role == RoleMerchant {
		_, err = psql.queries.CreateMerchant(ctx, db.CreateMerchantParams{
			UserID:    created.ID,
			StoreName: defaultStoreName(created.Email),
			Description: sql.NullString{
				Valid: false,
			},
		})
		if err != nil {
			return User{}, err
		}
	}
	return toUser(created), nil
}

func defaultStoreName(email string) string {
	localPart := strings.TrimSpace(strings.SplitN(email, "@", 2)[0])
	if localPart == "" {
		return "My Store"
	}
	return localPart + " Store"
}

func (psql *Store) GetUserByEmail(ctx context.Context, email string) (User, bool, error) {
	row, err := psql.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, false, nil
		}
		return User{}, false, err
	}
	return toUser(row), true, nil
}

func (psql *Store) GetUserByID(ctx context.Context, id int) (User, bool, error) {
	row, err := psql.queries.GetUserByID(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return User{}, false, nil
		}
		return User{}, false, err
	}
	return toUser(row), true, nil
}
