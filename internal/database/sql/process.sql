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

-- name: CountProcessingLogsFiltered :one
SELECT COUNT(*) as count
FROM tbl_processing_logs
WHERE (? = '' OR name LIKE '%' || ? || '%')
  AND (? = '' OR created_at >= ?)
  AND (? = '' OR created_at <= ?);

-- name: GetProcessingLogsFiltered :many
SELECT id, google_drive_url, name, template, start_date, end_date,
       excluded_rows, useragent, output_path, errors, created_at
FROM tbl_processing_logs
WHERE (? = '' OR name LIKE '%' || ? || '%')
  AND (? = '' OR created_at >= ?)
  AND (? = '' OR created_at <= ?)
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: GetProcessingLogByID :one
SELECT id, google_drive_url, name, template, start_date, end_date,
       excluded_rows, useragent, output_path, errors, created_at
FROM tbl_processing_logs
WHERE id = ?;
