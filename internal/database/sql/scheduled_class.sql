-- name: InsertScheduledClassSeries :one
INSERT INTO tbl_scheduled_class_series (created_by_role)
VALUES (?)
RETURNING id;

-- name: InsertScheduledClass :one
INSERT INTO tbl_scheduled_classes (student_id, teacher_id, scheduled_date, start_time, duration_minutes, rate, currency, is_trial_class, series_id, status, reason, created_by_role)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'scheduled', ?, ?)
RETURNING id;

-- name: GetScheduledClassByID :one
SELECT sc.id, sc.student_id, sc.teacher_id, sc.scheduled_date, sc.start_time, sc.duration_minutes, sc.rate, sc.currency, sc.is_trial_class, sc.series_id, sc.status, sc.reason, sc.created_by_role, sc.created_at, sc.updated_at,
	s.name as student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
JOIN tbl_teachers t ON sc.teacher_id = t.id
WHERE sc.id = ? AND sc.deleted_at IS NULL;

-- name: CountScheduledClassesBySeries :one
SELECT COUNT(*) as count
FROM tbl_scheduled_classes
WHERE series_id = ?
	AND deleted_at IS NULL
	AND status = 'scheduled';

-- name: CountScheduledClassesInSeriesFromDate :one
SELECT COUNT(*) as count
FROM tbl_scheduled_classes
WHERE series_id = ?
	AND scheduled_date >= ?
	AND deleted_at IS NULL
	AND status = 'scheduled';

-- name: ListActiveScheduledClassSeries :many
SELECT
	ser.id AS series_id,
	sc.id AS anchor_class_id,
	sc.student_id,
	sc.teacher_id,
	sc.scheduled_date AS anchor_date,
	sc.start_time,
	sc.duration_minutes,
	sc.rate,
	sc.currency,
	s.name AS student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) AS teacher_name,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.assigned_color AS teacher_assigned_color,
	t.profile_picture AS teacher_profile_picture,
	(
		SELECT CAST(MIN(sc2.scheduled_date) AS TEXT)
		FROM tbl_scheduled_classes sc2
		WHERE sc2.series_id = ser.id
			AND sc2.deleted_at IS NULL
			AND sc2.status = 'scheduled'
	) AS first_date,
	(
		SELECT CAST(MAX(sc2.scheduled_date) AS TEXT)
		FROM tbl_scheduled_classes sc2
		WHERE sc2.series_id = ser.id
			AND sc2.deleted_at IS NULL
			AND sc2.status = 'scheduled'
	) AS last_date,
	(
		SELECT COUNT(*)
		FROM tbl_scheduled_classes sc2
		WHERE sc2.series_id = ser.id
			AND sc2.deleted_at IS NULL
			AND sc2.status = 'scheduled'
	) AS scheduled_count
FROM tbl_scheduled_class_series ser
INNER JOIN tbl_scheduled_classes sc ON sc.series_id = ser.id
	AND sc.deleted_at IS NULL
	AND sc.status = 'scheduled'
	AND sc.id = (
		SELECT sc3.id
		FROM tbl_scheduled_classes sc3
		WHERE sc3.series_id = ser.id
			AND sc3.deleted_at IS NULL
			AND sc3.status = 'scheduled'
		ORDER BY
			CASE WHEN sc3.scheduled_date >= date('now', 'localtime') THEN 0 ELSE 1 END,
			sc3.scheduled_date ASC,
			sc3.start_time ASC,
			sc3.id ASC
		LIMIT 1
	)
INNER JOIN tbl_students s ON sc.student_id = s.id
INNER JOIN tbl_teachers t ON sc.teacher_id = t.id
WHERE EXISTS (
	SELECT 1
	FROM tbl_scheduled_classes sc4
	WHERE sc4.series_id = ser.id
		AND sc4.deleted_at IS NULL
		AND sc4.status = 'scheduled'
)
	AND (? = 0 OR sc.teacher_id = ?)
ORDER BY sc.scheduled_date ASC, sc.start_time ASC, ser.id ASC;

-- name: GetScheduledClassesInSeriesFromDate :many
SELECT
	sc.id,
	sc.student_id,
	sc.teacher_id,
	sc.scheduled_date,
	sc.start_time,
	sc.duration_minutes,
	sc.rate,
	sc.currency,
	sc.is_trial_class,
	sc.series_id,
	sc.status
FROM tbl_scheduled_classes sc
WHERE sc.series_id = ?
	AND sc.scheduled_date >= ?
	AND sc.deleted_at IS NULL
	AND sc.status = 'scheduled'
ORDER BY sc.scheduled_date ASC, sc.start_time ASC, sc.id ASC;

-- name: ListScheduledClassDatesBySeries :many
SELECT sc.scheduled_date
FROM tbl_scheduled_classes sc
WHERE sc.series_id = ?
	AND sc.deleted_at IS NULL
	AND sc.status = 'scheduled'
ORDER BY sc.scheduled_date ASC, sc.start_time ASC, sc.id ASC;

-- name: UpdateScheduledClassStatus :exec
UPDATE tbl_scheduled_classes
SET status = ?, reason = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL;

-- name: RescheduleScheduledClass :exec
UPDATE tbl_scheduled_classes
SET scheduled_date = ?, start_time = ?, status = 'scheduled', reason = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL;

-- name: UpdateScheduledClassDetails :exec
UPDATE tbl_scheduled_classes
SET student_id = ?, rate = ?, currency = ?, is_trial_class = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL;

-- name: UpdateScheduledClassSchedule :exec
UPDATE tbl_scheduled_classes
SET student_id = ?, scheduled_date = ?, start_time = ?, duration_minutes = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL;

-- name: SoftDeleteScheduledClass :exec
UPDATE tbl_scheduled_classes
SET reason = ?, deleted_at = datetime('now'), updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL;

-- name: CountScheduledClassesFiltered :one
SELECT COUNT(*) as count
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
WHERE (? = 0 OR sc.teacher_id = ?) AND sc.scheduled_date >= ? AND sc.scheduled_date <= ?
	AND sc.deleted_at IS NULL
	AND (? = '' OR sc.status = ?)
	AND (? = '' OR s.name LIKE '%' || ? || '%');

-- name: GetScheduledClassesFiltered :many
SELECT sc.id, sc.student_id, sc.teacher_id, sc.scheduled_date, sc.start_time, sc.duration_minutes, sc.rate, sc.currency, sc.status, sc.reason, sc.created_by_role, sc.created_at, sc.updated_at, sc.series_id,
	s.name as student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name,
	t.first_name as teacher_first_name, t.middle_name as teacher_middle_name, t.last_name as teacher_last_name,
	t.assigned_color as teacher_assigned_color, t.profile_picture as teacher_profile_picture
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
JOIN tbl_teachers t ON sc.teacher_id = t.id
WHERE (? = 0 OR sc.teacher_id = ?) AND sc.scheduled_date >= ? AND sc.scheduled_date <= ?
	AND sc.deleted_at IS NULL
	AND (? = '' OR sc.status = ?)
	AND (? = '' OR s.name LIKE '%' || ? || '%')
ORDER BY CASE WHEN sc.scheduled_date = date('now', 'localtime') THEN 0 ELSE 1 END, sc.scheduled_date ASC, sc.start_time ASC, sc.created_at ASC
LIMIT ? OFFSET ?;

-- name: CountScheduledDuplicate :one
SELECT COUNT(*) as count
FROM tbl_scheduled_classes
WHERE student_id = ? AND teacher_id = ? AND scheduled_date = ? AND duration_minutes = ?
	AND status = 'scheduled'
	AND deleted_at IS NULL
	AND (? = 0 OR id != ?);

-- name: GetScheduledDuplicate :one
SELECT
	sc.id,
	sc.scheduled_date,
	sc.start_time,
	sc.duration_minutes,
	sc.status,
	s.name as student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
JOIN tbl_teachers t ON sc.teacher_id = t.id
WHERE sc.student_id = ? AND sc.teacher_id = ? AND sc.scheduled_date = ? AND sc.duration_minutes = ?
	AND sc.status = 'scheduled'
	AND sc.deleted_at IS NULL
	AND (? = 0 OR sc.id != ?)
LIMIT 1;

-- name: CountScheduledClassesByStatusAndDate :one
SELECT COUNT(*) as count
FROM tbl_scheduled_classes
WHERE scheduled_date = ? AND status = ?
	AND deleted_at IS NULL
	AND (? = 0 OR teacher_id = ?);

-- name: GetScheduledClassesByTeacherOnDate :many
SELECT
	sc.id,
	sc.scheduled_date,
	sc.start_time,
	sc.duration_minutes,
	sc.status,
	s.name as student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
JOIN tbl_teachers t ON sc.teacher_id = t.id
WHERE sc.teacher_id = ? AND sc.scheduled_date = ? AND sc.status = 'scheduled'
	AND sc.deleted_at IS NULL
	AND sc.start_time IS NOT NULL AND trim(sc.start_time) != ''
	AND (? = 0 OR sc.id != ?);

-- name: GetScheduledClassesByStudentOnDate :many
SELECT
	sc.id,
	sc.scheduled_date,
	sc.start_time,
	sc.duration_minutes,
	sc.status,
	s.name as student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
JOIN tbl_teachers t ON sc.teacher_id = t.id
WHERE sc.student_id = ? AND sc.scheduled_date = ? AND sc.status = 'scheduled'
	AND sc.deleted_at IS NULL
	AND sc.start_time IS NOT NULL AND trim(sc.start_time) != ''
	AND (? = 0 OR sc.id != ?);
