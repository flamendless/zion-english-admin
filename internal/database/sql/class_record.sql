-- name: InsertClassRecord :exec
INSERT INTO tbl_class_records (student_id, teacher_id, date, duration_minutes, rate, currency, status, reason, notes, recorded_by_role)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetClassRecordsByTeacherAndDateRange :many
SELECT cr.id, cr.student_id, cr.teacher_id, cr.date, cr.duration_minutes, cr.rate, cr.currency, cr.status, cr.reason, cr.notes, cr.created_at, cr.updated_at, cr.recorded_by_role,
       s.name as student_name, t.name as teacher_name
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE cr.teacher_id = ? AND cr.date >= ? AND cr.date <= ?
ORDER BY cr.created_at DESC;

-- name: GetTotalRateByTeacherAndDateRange :one
SELECT COALESCE(SUM(cr.rate), 0) as total_rate
FROM tbl_class_records cr
WHERE (? = 0 OR cr.teacher_id = ?) AND cr.date >= ? AND cr.date <= ? AND cr.status = 'conducted';

-- name: GetClassRecordByID :one
SELECT cr.id, cr.student_id, cr.teacher_id, cr.date, cr.duration_minutes, cr.rate, cr.currency, cr.status, cr.reason, cr.notes, cr.created_at, cr.updated_at, cr.recorded_by_role,
       s.name as student_name, t.name as teacher_name
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE cr.id = ?;

-- name: UpdateClassRecord :exec
UPDATE tbl_class_records
SET student_id = ?, teacher_id = ?, date = ?, duration_minutes = ?, rate = ?, currency = ?, status = ?, reason = ?, notes = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: CountClassRecordDuplicate :one
SELECT COUNT(*) as count
FROM tbl_class_records
WHERE student_id = ? AND teacher_id = ? AND date = ? AND duration_minutes = ?
  AND (? = 0 OR id != ?);

-- name: CountClassRecordsFiltered :one
SELECT COUNT(*) as count
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
WHERE (? = 0 OR cr.teacher_id = ?) AND cr.date >= ? AND cr.date <= ?
  AND (? = '' OR cr.status = ?)
  AND (? = '' OR s.name LIKE '%' || ? || '%');

-- name: GetClassRecordsFiltered :many
SELECT cr.id, cr.student_id, cr.teacher_id, cr.date, cr.duration_minutes, cr.rate, cr.currency, cr.status, cr.reason, cr.notes, cr.created_at, cr.updated_at, cr.recorded_by_role,
       s.name as student_name, t.name as teacher_name
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE (? = 0 OR cr.teacher_id = ?) AND cr.date >= ? AND cr.date <= ?
  AND (? = '' OR cr.status = ?)
  AND (? = '' OR s.name LIKE '%' || ? || '%')
ORDER BY CASE WHEN cr.date = date('now', 'localtime') THEN 0 ELSE 1 END, cr.date DESC, cr.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountClassRecordsByStatusAndDateRange :many
SELECT cr.status, COUNT(*) as count
FROM tbl_class_records cr
WHERE cr.date >= ? AND cr.date <= ?
  AND (? = 0 OR cr.teacher_id = ?)
GROUP BY cr.status;

-- name: SumConductedRateByCurrencyAndDateRange :many
SELECT cr.currency, COALESCE(SUM(cr.rate), 0) as total_rate
FROM tbl_class_records cr
WHERE cr.date >= ? AND cr.date <= ? AND cr.status = 'conducted'
  AND (? = 0 OR cr.teacher_id = ?)
GROUP BY cr.currency;
