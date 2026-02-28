package database

import (
	"context"

	"zion-english/internal/database/queries"
)

type ProcessingLog = queries.TblProcessingLog

func InsertProcessingLog(db Service, log *ProcessingLog) error {
	params := queries.InsertProcessingLogParams{
		GoogleDriveUrl: log.GoogleDriveUrl,
		Name:           log.Name,
		Template:       log.Template,
		StartDate:      log.StartDate,
		EndDate:        log.EndDate,
		ExcludedRows:   log.ExcludedRows,
		Useragent:      log.Useragent,
		OutputPath:     log.OutputPath,
		Errors:         log.Errors,
	}
	return db.GetQueries().InsertProcessingLog(context.Background(), params)
}

func GetAllProcessingLogs(db Service) ([]ProcessingLog, error) {
	return db.GetQueries().GetAllProcessingLogs(context.Background())
}

func GetProcessingLogByID(db Service, id int64) (*ProcessingLog, error) {
	log, err := db.GetQueries().GetProcessingLogByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return &log, nil
}
