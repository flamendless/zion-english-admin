package database

import (
	"context"
	"database/sql"

	"zion-english/internal/database/queries"
)

type SystemLog = queries.GetAllLogsRow

func GetAllSystemLogs(db Service) ([]SystemLog, error) {
	return db.GetQueries().GetAllLogs(context.Background())
}

func GetSystemLogsByCreatedBy(db Service, teacherID int64) ([]SystemLog, error) {
	rows, err := db.GetQueries().GetLogsByCreatedBy(context.Background(), sql.NullInt64{Int64: teacherID, Valid: true})
	if err != nil {
		return nil, err
	}
	logs := make([]SystemLog, len(rows))
	for i, row := range rows {
		logs[i] = SystemLog(row)
	}
	return logs, nil
}
