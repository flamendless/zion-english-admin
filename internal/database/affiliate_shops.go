package database

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

func EnsureAffiliatedProductShopID(ctx context.Context, db Service, brandName string) (sql.NullInt64, error) {
	brandName = strings.TrimSpace(brandName)
	if brandName == "" {
		return sql.NullInt64{}, nil
	}

	q := db.GetQueries()
	row, err := q.GetAffiliatedProductShopByBrandName(ctx, brandName)
	if err == nil {
		return sql.NullInt64{Int64: row.ID, Valid: true}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return sql.NullInt64{}, err
	}

	id, err := q.InsertAffiliatedProductShop(ctx, brandName)
	if err != nil {
		if IsUniqueConstraint(err) {
			row, err := q.GetAffiliatedProductShopByBrandName(ctx, brandName)
			if err != nil {
				return sql.NullInt64{}, err
			}
			return sql.NullInt64{Int64: row.ID, Valid: true}, nil
		}
		return sql.NullInt64{}, err
	}
	return sql.NullInt64{Int64: id, Valid: true}, nil
}
