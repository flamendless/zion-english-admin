package database

import (
	"time"
)

type ProcessingLog struct {
	ID             int64
	GoogleDriveURL string
	Name           string
	Template       string
	StartDate      string
	EndDate        string
	ExcludedRows   string
	UserAgent      string
	OutputPath     string
	Errors         string
	CreatedAt      time.Time
}

func InsertProcessingLog(log *ProcessingLog) (int64, error) {
	query := `
		INSERT INTO processing_logs (
			google_drive_url, name, template, start_date, end_date,
			excluded_rows, useragent, output_path, errors
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := DB.Exec(
		query,
		log.GoogleDriveURL,
		log.Name,
		log.Template,
		log.StartDate,
		log.EndDate,
		log.ExcludedRows,
		log.UserAgent,
		log.OutputPath,
		log.Errors,
	)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func GetAllProcessingLogs() ([]ProcessingLog, error) {
	query := `
		SELECT id, google_drive_url, name, template, start_date, end_date,
		       excluded_rows, useragent, output_path, errors, created_at
		FROM processing_logs
		ORDER BY created_at DESC
	`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ProcessingLog
	for rows.Next() {
		var log ProcessingLog
		err := rows.Scan(
			&log.ID,
			&log.GoogleDriveURL,
			&log.Name,
			&log.Template,
			&log.StartDate,
			&log.EndDate,
			&log.ExcludedRows,
			&log.UserAgent,
			&log.OutputPath,
			&log.Errors,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func GetProcessingLogByID(id int64) (*ProcessingLog, error) {
	query := `
		SELECT id, google_drive_url, name, template, start_date, end_date,
		       excluded_rows, useragent, output_path, errors, created_at
		FROM processing_logs
		WHERE id = ?
	`

	var log ProcessingLog
	err := DB.QueryRow(query, id).Scan(
		&log.ID,
		&log.GoogleDriveURL,
		&log.Name,
		&log.Template,
		&log.StartDate,
		&log.EndDate,
		&log.ExcludedRows,
		&log.UserAgent,
		&log.OutputPath,
		&log.Errors,
		&log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &log, nil
}
