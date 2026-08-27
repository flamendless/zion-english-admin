-- name: InsertStudent :one
INSERT INTO tbl_students (name, currency, contact, rate_per_class, parent_name, assigned_color, status, inactive_reason)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: GetAllStudents :many
SELECT id, name, currency, contact, rate_per_class, parent_name, assigned_color, status, inactive_reason, created_at, updated_at
FROM tbl_students
ORDER BY created_at DESC;

-- name: GetActiveStudents :many
SELECT id, name, currency, contact, rate_per_class, parent_name, assigned_color, status, inactive_reason, created_at, updated_at
FROM tbl_students
WHERE status = 'active'
ORDER BY name ASC;

-- name: GetStudentByID :one
SELECT id, name, currency, contact, rate_per_class, parent_name, assigned_color, status, inactive_reason, created_at, updated_at
FROM tbl_students
WHERE id = ?;

-- name: UpdateStudent :exec
UPDATE tbl_students
SET name = ?, currency = ?, contact = ?, rate_per_class = ?, parent_name = ?, assigned_color = ?, status = ?, inactive_reason = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: SearchStudentsByName :many
SELECT id, name, currency, contact, rate_per_class, parent_name, assigned_color, status, inactive_reason, created_at, updated_at
FROM tbl_students
WHERE name LIKE '%' || ? || '%'
ORDER BY name ASC
LIMIT 10;

-- name: CountStudentsFiltered :one
SELECT COUNT(DISTINCT s.id) as count
FROM tbl_students s
LEFT JOIN tbl_teachers_students_m2m m2m ON s.id = m2m.student_id
WHERE (? = '' OR s.name LIKE '%' || ? || '%')
	AND (? = '' OR s.status = ?)
	AND (? = 0 OR m2m.teacher_id = ?);

-- name: GetStudentsFiltered :many
SELECT DISTINCT s.id, s.name, s.currency, s.contact, s.rate_per_class, s.parent_name, s.assigned_color, s.status, s.inactive_reason, s.created_at, s.updated_at
FROM tbl_students s
LEFT JOIN tbl_teachers_students_m2m m2m ON s.id = m2m.student_id
WHERE (? = '' OR s.name LIKE '%' || ? || '%')
	AND (? = '' OR s.status = ?)
	AND (? = 0 OR m2m.teacher_id = ?)
ORDER BY s.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountStudentsByStatus :many
SELECT status, COUNT(*) as count
FROM tbl_students
GROUP BY status;
