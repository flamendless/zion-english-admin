-- name: InsertTeacher :exec
INSERT INTO tbl_teachers (name, birthdate, address, joining_date, mobile_number, email, certifications, assigned_color, rate_per_class, currency, drive_url, sex, password, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetTeacherByName :one
SELECT COUNT(*) as count FROM tbl_teachers WHERE name = ?;

-- name: GetTeacherCountByEmail :one
SELECT COUNT(*) as count FROM tbl_teachers WHERE email = ?;

-- name: GetTeacherIDByName :one
SELECT id FROM tbl_teachers WHERE name = ?;

-- name: GetTeacherByEmail :one
SELECT id, name, email, password, status FROM tbl_teachers WHERE email = ?;

-- name: GetAllTeachers :many
SELECT id, name, birthdate, address, joining_date, mobile_number, email, certifications, assigned_color, rate_per_class, currency, drive_url, sex, password, template, created_at, updated_at, status
FROM tbl_teachers
ORDER BY CASE status WHEN 'pending' THEN 0 ELSE 1 END, name ASC;

-- name: GetApprovedTeachers :many
SELECT id, name, birthdate, address, joining_date, mobile_number, email, certifications, assigned_color, rate_per_class, currency, drive_url, sex, password, template, created_at, updated_at, status
FROM tbl_teachers
WHERE status = 'approved'
ORDER BY name ASC;

-- name: ApproveTeacher :exec
UPDATE tbl_teachers SET status = 'approved', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'pending';

-- name: UnapproveTeacher :exec
UPDATE tbl_teachers SET status = 'pending', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'approved';

-- name: GetTeacherByID :one
SELECT id, name, template FROM tbl_teachers WHERE id = ?;

-- name: UpdateTeacherTemplate :exec
UPDATE tbl_teachers SET template = ? WHERE id = ?;

-- name: UpdateTeacherPassword :exec
UPDATE tbl_teachers SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?;
