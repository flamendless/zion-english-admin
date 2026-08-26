-- name: InsertScheduledClass :one
INSERT INTO tbl_scheduled_classes (student_id, teacher_id, scheduled_date, start_time, duration_minutes, rate, currency, status, reason, created_by_role)
VALUES (?, ?, ?, ?, ?, ?, ?, 'scheduled', ?, ?)
RETURNING id;

-- name: GetScheduledClassByID :one
SELECT sc.id, sc.student_id, sc.teacher_id, sc.scheduled_date, sc.start_time, sc.duration_minutes, sc.rate, sc.currency, sc.status, sc.reason, sc.created_by_role, sc.created_at, sc.updated_at,
       s.name as student_name,
       trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
JOIN tbl_teachers t ON sc.teacher_id = t.id
WHERE sc.id = ?;

-- name: UpdateScheduledClassStatus :exec
UPDATE tbl_scheduled_classes
SET status = ?, reason = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: RescheduleScheduledClass :exec
UPDATE tbl_scheduled_classes
SET scheduled_date = ?, start_time = ?, status = 'scheduled', reason = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: UpdateScheduledClassDetails :exec
UPDATE tbl_scheduled_classes
SET student_id = ?, rate = ?, currency = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: CountScheduledClassesFiltered :one
SELECT COUNT(*) as count
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
WHERE (? = 0 OR sc.teacher_id = ?) AND sc.scheduled_date >= ? AND sc.scheduled_date <= ?
  AND (? = '' OR sc.status = ?)
  AND (? = '' OR s.name LIKE '%' || ? || '%');

-- name: GetScheduledClassesFiltered :many
SELECT sc.id, sc.student_id, sc.teacher_id, sc.scheduled_date, sc.start_time, sc.duration_minutes, sc.rate, sc.currency, sc.status, sc.reason, sc.created_by_role, sc.created_at, sc.updated_at,
       s.name as student_name,
       trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name,
       t.first_name as teacher_first_name, t.middle_name as teacher_middle_name, t.last_name as teacher_last_name,
       t.assigned_color as teacher_assigned_color, t.profile_picture as teacher_profile_picture
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
JOIN tbl_teachers t ON sc.teacher_id = t.id
WHERE (? = 0 OR sc.teacher_id = ?) AND sc.scheduled_date >= ? AND sc.scheduled_date <= ?
  AND (? = '' OR sc.status = ?)
  AND (? = '' OR s.name LIKE '%' || ? || '%')
ORDER BY CASE WHEN sc.scheduled_date = date('now', 'localtime') THEN 0 ELSE 1 END, sc.scheduled_date ASC, sc.start_time ASC, sc.created_at ASC
LIMIT ? OFFSET ?;

-- name: CountScheduledDuplicate :one
SELECT COUNT(*) as count
FROM tbl_scheduled_classes
WHERE student_id = ? AND teacher_id = ? AND scheduled_date = ? AND duration_minutes = ?
  AND status = 'scheduled'
  AND (? = 0 OR id != ?);

-- name: CountScheduledClassesByStatusAndDate :one
SELECT COUNT(*) as count
FROM tbl_scheduled_classes
WHERE scheduled_date = ? AND status = ?
  AND (? = 0 OR teacher_id = ?);

-- name: GetScheduledClassesByTeacherOnDate :many
SELECT id, start_time, duration_minutes
FROM tbl_scheduled_classes
WHERE teacher_id = ? AND scheduled_date = ? AND status = 'scheduled'
  AND start_time IS NOT NULL AND trim(start_time) != ''
  AND (? = 0 OR id != ?);

-- name: GetScheduledClassesByStudentOnDate :many
SELECT id, start_time, duration_minutes
FROM tbl_scheduled_classes
WHERE student_id = ? AND scheduled_date = ? AND status = 'scheduled'
  AND start_time IS NOT NULL AND trim(start_time) != ''
  AND (? = 0 OR id != ?);
