-- name: InsertStudent :exec
INSERT INTO tbl_students (name, currency, contact, rate_per_class, parent_name, assigned_color, status)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetStudentByName :one
SELECT COUNT(*) as count FROM tbl_students WHERE name = ?;

-- name: GetStudentIDByName :one
SELECT id FROM tbl_students WHERE name = ?;

-- name: GetAllStudents :many
SELECT id, name, currency, contact, rate_per_class, parent_name, assigned_color, status, created_at, updated_at
FROM tbl_students
ORDER BY created_at DESC;

-- name: GetActiveStudents :many
SELECT id, name, currency, contact, rate_per_class, parent_name, assigned_color, status, created_at, updated_at
FROM tbl_students
WHERE status = 'active'
ORDER BY name ASC;

-- name: GetStudentByID :one
SELECT id, name, currency, contact, rate_per_class, parent_name, assigned_color, status, created_at, updated_at
FROM tbl_students
WHERE id = ?;

-- name: SearchStudentsByName :many
SELECT id, name, currency, contact, rate_per_class, parent_name, assigned_color, status, created_at, updated_at
FROM tbl_students
WHERE name LIKE '%' || ? || '%'
ORDER BY name ASC
LIMIT 10;

