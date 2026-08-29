package featureflags

import (
	"context"
	"database/sql"
	"errors"
	"zion-english/internal/constants"
	"zion-english/internal/database"
	"zion-english/internal/database/queries"
)

func IsEnabled(ctx context.Context, db database.Service, key constants.FeatureFlagKey) bool {
	row, err := db.GetQueries().GetFeatureFlag(ctx, string(key))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true
		}
		return true
	}
	return row.Enabled == 1
}

func SetEnabled(ctx context.Context, db database.Service, key constants.FeatureFlagKey, enabled bool) error {
	var enabledInt int64
	if enabled {
		enabledInt = 1
	}
	return db.GetQueries().UpsertFeatureFlag(ctx, queries.UpsertFeatureFlagParams{
		Key:     string(key),
		Enabled: enabledInt,
	})
}
