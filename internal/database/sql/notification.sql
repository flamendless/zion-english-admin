-- name: GetNotificationForSuperuser :one
SELECT id, from_teacher_id, from_name, to_teacher_id, to_name, message, kind, dedupe_key, read, read_at, created_at
FROM tbl_notifications
WHERE id = ? AND to_teacher_id IS NULL;

-- name: GetNotificationForTeacher :one
SELECT id, from_teacher_id, from_name, to_teacher_id, to_name, message, kind, dedupe_key, read, read_at, created_at
FROM tbl_notifications
WHERE id = ? AND to_teacher_id = ?;

-- name: InsertNotification :exec
INSERT OR IGNORE INTO tbl_notifications (
	from_teacher_id,
	from_name,
	to_teacher_id,
	to_name,
	message,
	kind,
	dedupe_key,
	read,
	created_at
) VALUES (
	?, ?, ?, ?, ?, ?, ?, 0, datetime('now')
);

-- name: CountUnreadNotificationsForSuperuser :one
SELECT COUNT(*) FROM tbl_notifications
WHERE read = 0 AND to_teacher_id IS NULL;

-- name: CountUnreadNotificationsForTeacher :one
SELECT COUNT(*) FROM tbl_notifications
WHERE read = 0 AND to_teacher_id = ?;

-- name: CountNotificationsForSuperuser :one
SELECT COUNT(*) FROM tbl_notifications
WHERE to_teacher_id IS NULL
	AND (? = 0 OR read = 0);

-- name: CountNotificationsForTeacher :one
SELECT COUNT(*) FROM tbl_notifications
WHERE to_teacher_id = ?
	AND (? = 0 OR read = 0);

-- name: GetNotificationsPagedForSuperuser :many
SELECT id, from_teacher_id, from_name, to_teacher_id, to_name, message, kind, dedupe_key, read, read_at, created_at
FROM tbl_notifications
WHERE to_teacher_id IS NULL
	AND (? = 0 OR read = 0)
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: GetNotificationsPagedForTeacher :many
SELECT id, from_teacher_id, from_name, to_teacher_id, to_name, message, kind, dedupe_key, read, read_at, created_at
FROM tbl_notifications
WHERE to_teacher_id = ?
	AND (? = 0 OR read = 0)
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: GetRecentNotificationsForSuperuser :many
SELECT id, from_teacher_id, from_name, to_teacher_id, to_name, message, kind, dedupe_key, read, read_at, created_at
FROM tbl_notifications
WHERE to_teacher_id IS NULL
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: GetRecentNotificationsForTeacher :many
SELECT id, from_teacher_id, from_name, to_teacher_id, to_name, message, kind, dedupe_key, read, read_at, created_at
FROM tbl_notifications
WHERE to_teacher_id = ?
ORDER BY created_at DESC, id DESC
LIMIT ?;

-- name: MarkNotificationReadForSuperuser :exec
UPDATE tbl_notifications
SET read = 1, read_at = datetime('now')
WHERE id = ? AND to_teacher_id IS NULL AND read = 0;

-- name: MarkNotificationReadForTeacher :exec
UPDATE tbl_notifications
SET read = 1, read_at = datetime('now')
WHERE id = ? AND to_teacher_id = ? AND read = 0;

-- name: MarkAllNotificationsReadForSuperuser :exec
UPDATE tbl_notifications
SET read = 1, read_at = datetime('now')
WHERE to_teacher_id IS NULL AND read = 0;

-- name: MarkAllNotificationsReadForTeacher :exec
UPDATE tbl_notifications
SET read = 1, read_at = datetime('now')
WHERE to_teacher_id = ? AND read = 0;

-- name: GetMissedScheduledClasses :many
SELECT
	sc.id,
	sc.teacher_id,
	sc.scheduled_date,
	sc.start_time,
	sc.duration_minutes,
	s.name AS student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) AS teacher_name
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON s.id = sc.student_id
JOIN tbl_teachers t ON t.id = sc.teacher_id
WHERE sc.status = 'scheduled'
	AND sc.deleted_at IS NULL
	AND datetime(sc.scheduled_date || ' ' || COALESCE(sc.start_time, '00:00'), '+' || sc.duration_minutes || ' minutes') < ?;
