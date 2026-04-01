-- name: InsertTeacher :exec
INSERT INTO tbl_teachers (name, birthdate, address, joining_date, mobile_number, email, certifications, assigned_color, rate_per_class, currency, drive_url, sex, password)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTeacherByName :one
SELECT COUNT(*) as count FROM tbl_teachers WHERE name = ?;

-- name: GetTeacherByEmail :one
SELECT id, name, email, password FROM tbl_teachers WHERE email = ?;

-- name: GetAllTeachers :many
SELECT id, name, birthdate, address, joining_date, mobile_number, email, certifications, assigned_color, rate_per_class, currency, drive_url, sex, password, created_at, updated_at
FROM tbl_teachers
ORDER BY name ASC;

