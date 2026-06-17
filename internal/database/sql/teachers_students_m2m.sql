-- name: InsertTeacherStudentM2M :exec
INSERT INTO tbl_teachers_students_m2m (teacher_id, student_id)
VALUES (?, ?);

-- name: IsStudentAssignedToTeacher :one
SELECT COUNT(*) as count FROM tbl_teachers_students_m2m WHERE teacher_id = ? AND student_id = ?;

-- name: GetStudentsByTeacherID :many
SELECT s.id, s.name, s.currency, s.contact, s.rate_per_class, s.parent_name, s.assigned_color, s.status, s.created_at, s.updated_at
FROM tbl_students s
INNER JOIN tbl_teachers_students_m2m m2m ON s.id = m2m.student_id
WHERE m2m.teacher_id = ?
ORDER BY s.name ASC;

-- name: GetTeachersByStudentID :many
SELECT t.id, t.name, t.birthdate, t.address, t.joining_date, t.mobile_number, t.email, t.certifications, t.assigned_color, t.rate_per_class, t.currency, t.drive_url, t.sex, t.password, t.created_at, t.updated_at
FROM tbl_teachers t
INNER JOIN tbl_teachers_students_m2m m2m ON t.id = m2m.teacher_id
WHERE m2m.student_id = ?
ORDER BY t.name ASC;
