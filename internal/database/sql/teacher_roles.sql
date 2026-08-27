-- name: GetTeacherRolesByTeacherID :many
SELECT role
FROM tbl_teacher_roles
WHERE teacher_id = ?
ORDER BY role ASC;

-- name: GetTeacherRolesByTeacherIDs :many
SELECT teacher_id, role
FROM tbl_teacher_roles
WHERE teacher_id IN (sqlc.slice('teacher_ids'))
ORDER BY teacher_id ASC, role ASC;

-- name: InsertTeacherRole :exec
INSERT INTO tbl_teacher_roles (teacher_id, role)
VALUES (?, ?);

-- name: DeleteTeacherRole :exec
DELETE FROM tbl_teacher_roles
WHERE teacher_id = ? AND role = ?;

-- name: DeleteTeacherRolesByTeacherID :exec
DELETE FROM tbl_teacher_roles
WHERE teacher_id = ?;

-- name: TeacherHasRole :one
SELECT COUNT(*) AS count
FROM tbl_teacher_roles
WHERE teacher_id = ? AND role = ?;
