package database

import (
	"context"

	"zion-english/internal/database/queries"
)

type SystemLog = queries.GetAllLogsRow

func GetAllSystemLogs(db Service) ([]SystemLog, error) {
	return db.GetQueries().GetAllLogs(context.Background())
}
