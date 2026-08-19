package reports

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
)

type FingerprintRow struct {
	ID              int64
	StudentID       int64
	Date            string
	StartTime       sql.NullString
	EndTime         sql.NullString
	DurationMinutes int64
	Rate            float64
	Currency        string
	Status          string
	UpdatedAt       string
}

func Fingerprint(rows []FingerprintRow) string {
	h := sha256.New()
	for _, row := range rows {
		fmt.Fprintf(h, "%d|%d|%s|%s|%s|%d|%.4f|%s|%s|%s\n",
			row.ID,
			row.StudentID,
			row.Date,
			nullString(row.StartTime),
			nullString(row.EndTime),
			row.DurationMinutes,
			row.Rate,
			row.Currency,
			row.Status,
			row.UpdatedAt,
		)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func nullString(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}
