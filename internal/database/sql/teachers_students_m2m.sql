-- name: InsertTeacherStudentM2M :exec
INSERT INTO tbl_teachers_students_m2m (teacher_id, student_id)
VALUES (?, ?);

-- name: IsStudentAssignedToTeacher :one
SELECT COUNT(*) as count FROM tbl_teachers_students_m2m WHERE teacher_id = ? AND student_id = ?;

-- name: DeleteTeacherStudentLinksByStudentID :exec
DELETE FROM tbl_teachers_students_m2m WHERE student_id = ?;

-- name: GetStudentsByTeacherID :many
SELECT s.id, s.name, s.currency, s.contact, s.rate_per_class, s.parent_name, s.assigned_color, s.status, s.created_at, s.updated_at
FROM tbl_students s
INNER JOIN tbl_teachers_students_m2m m2m ON s.id = m2m.student_id
WHERE m2m.teacher_id = ?
ORDER BY s.name ASC;

-- name: CountStudentsByTeacherIDFiltered :one
SELECT COUNT(*) as count
FROM tbl_students s
INNER JOIN tbl_teachers_students_m2m m2m ON s.id = m2m.student_id
WHERE m2m.teacher_id = ?
	AND (? = '' OR s.name LIKE '%' || ? || '%')
	AND (? = '' OR s.status = ?);

-- name: GetStudentsByTeacherIDFiltered :many
SELECT s.id, s.name, s.currency, s.contact, s.rate_per_class, s.parent_name, s.assigned_color, s.status, s.created_at, s.updated_at
FROM tbl_students s
INNER JOIN tbl_teachers_students_m2m m2m ON s.id = m2m.student_id
WHERE m2m.teacher_id = ?
	AND (? = '' OR s.name LIKE '%' || ? || '%')
	AND (? = '' OR s.status = ?)
ORDER BY s.name ASC
LIMIT ? OFFSET ?;

-- name: CountStudentsByTeacherID :one
SELECT COUNT(*) as count
FROM tbl_teachers_students_m2m
WHERE teacher_id = ?;

-- name: GetTeachersByStudentID :many
SELECT t.id, t.first_name, t.middle_name, t.last_name, t.birthdate, t.address, t.joining_date, t.mobile_number, t.email, t.certifications, t.assigned_color, t.rate_per_class, t.currency, t.drive_url, t.sex, t.password, t.created_at, t.updated_at
FROM tbl_teachers t
INNER JOIN tbl_teachers_students_m2m m2m ON t.id = m2m.teacher_id
WHERE m2m.student_id = ?
ORDER BY t.last_name ASC, t.first_name ASC, t.middle_name ASC;

-- name: GetAllStudentTeacherNames :many
SELECT m2m.student_id,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name
FROM tbl_teachers_students_m2m m2m
INNER JOIN tbl_teachers t ON t.id = m2m.teacher_id
WHERE t.deleted = 0
ORDER BY m2m.student_id ASC, t.last_name ASC, t.first_name ASC, t.middle_name ASC;
