-- name: InsertRecord :exec
INSERT INTO tbl_records (google_drive_url, student_name, date, duration_minutes, rate, status)
VALUES (?, ?, ?, ?, ?, ?);

