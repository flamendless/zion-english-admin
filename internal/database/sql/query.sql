-- name: InsertProcessingLog :exec
INSERT INTO tbl_processing_logs (
    google_drive_url, name, template, start_date, end_date,
    excluded_rows, useragent, output_path, errors
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetAllProcessingLogs :many
SELECT id, google_drive_url, name, template, start_date, end_date,
       excluded_rows, useragent, output_path, errors, created_at
FROM tbl_processing_logs
ORDER BY created_at DESC;

-- name: GetProcessingLogByID :one
SELECT id, google_drive_url, name, template, start_date, end_date,
       excluded_rows, useragent, output_path, errors, created_at
FROM tbl_processing_logs
WHERE id = ?;

-- name: InsertRecord :exec
INSERT INTO tbl_records (google_drive_url, student_name, date, duration_minutes, rate, status)
VALUES (?, ?, ?, ?, ?, ?);
