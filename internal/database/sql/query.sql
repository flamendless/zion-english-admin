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

-- name: InsertStudent :exec
INSERT INTO tbl_students (name, currency, contact, rate_per_class, parent_name, assigned_color, status)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetStudentByName :one
SELECT COUNT(*) as count FROM tbl_students WHERE name = ?;

-- name: GetAllStudents :many
SELECT id, name, currency, contact, rate_per_class, parent_name, assigned_color, status, created_at, updated_at
FROM tbl_students
ORDER BY created_at DESC;

-- name: InsertTeacher :exec
INSERT INTO tbl_teachers (name, birthdate, address, joining_date, mobile_number, email, certifications, assigned_color, rate_per_class, currency, drive_url, sex)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTeacherByName :one
SELECT COUNT(*) as count FROM tbl_teachers WHERE name = ?;

-- name: GetAllTeachers :many
SELECT id, name, birthdate, address, joining_date, mobile_number, email, certifications, assigned_color, rate_per_class, currency, drive_url, sex, created_at, updated_at
FROM tbl_teachers
ORDER BY created_at DESC;

