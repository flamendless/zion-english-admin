-- name: InsertClassRecord :exec
INSERT INTO tbl_class_records (student_id, teacher_id, date, duration_minutes, rate, currency, status, reason, recorded_by_role)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetClassRecordsByTeacherAndDateRange :many
SELECT cr.id, cr.student_id, cr.teacher_id, cr.date, cr.duration_minutes, cr.rate, cr.currency, cr.status, cr.reason, cr.created_at, cr.updated_at, cr.recorded_by_role,
       s.name as student_name, t.name as teacher_name
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE cr.teacher_id = ? AND cr.date >= ? AND cr.date <= ?
ORDER BY cr.created_at DESC;

-- name: GetTotalRateByTeacherAndDateRange :one
SELECT COALESCE(SUM(cr.rate), 0) as total_rate
FROM tbl_class_records cr
WHERE cr.teacher_id = ? AND cr.date >= ? AND cr.date <= ?;

