-- name: InsertClassRecord :one
INSERT INTO tbl_class_records (student_id, teacher_id, date, start_time, end_time, duration_minutes, rate, currency, is_trial_class, status, reason, notes, recorded_by_role)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: GetClassRecordsByTeacherAndDateRange :many
SELECT cr.id, cr.student_id, cr.teacher_id, cr.date, cr.duration_minutes, cr.rate, cr.currency, cr.status, cr.reason, cr.notes, cr.created_at, cr.updated_at, cr.recorded_by_role,
	s.name as student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE cr.teacher_id = ? AND cr.date >= ? AND cr.date <= ?
	AND cr.deleted_at IS NULL
ORDER BY cr.created_at DESC;

-- name: GetTotalRateByTeacherAndDateRange :one
SELECT COALESCE(SUM(cr.rate), 0) as total_rate
FROM tbl_class_records cr
WHERE (? = 0 OR cr.teacher_id = ?) AND cr.date >= ? AND cr.date <= ? AND cr.status = 'conducted'
	AND cr.deleted_at IS NULL;

-- name: GetClassRecordByID :one
SELECT cr.id, cr.student_id, cr.teacher_id, cr.date, cr.start_time, cr.end_time, cr.duration_minutes, cr.rate, cr.currency, cr.is_trial_class, cr.status, cr.reason, cr.notes, cr.created_at, cr.updated_at, cr.recorded_by_role,
	s.name as student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name,
	t.first_name as teacher_first_name,
	t.middle_name as teacher_middle_name,
	t.last_name as teacher_last_name,
	t.assigned_color as teacher_assigned_color,
	t.profile_picture as teacher_profile_picture
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE cr.id = ? AND cr.deleted_at IS NULL;

-- name: UpdateClassRecord :exec
UPDATE tbl_class_records
SET student_id = ?, teacher_id = ?, date = ?, start_time = ?, end_time = ?, duration_minutes = ?, rate = ?, currency = ?, is_trial_class = ?, status = ?, reason = ?, notes = ?, updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL;

-- name: SoftDeleteClassRecord :exec
UPDATE tbl_class_records
SET reason = ?, deleted_at = datetime('now'), updated_at = datetime('now')
WHERE id = ? AND deleted_at IS NULL;

-- name: CountClassRecordDuplicate :one
SELECT COUNT(*) as count
FROM tbl_class_records
WHERE student_id = ? AND teacher_id = ? AND date = ? AND duration_minutes = ?
	AND deleted_at IS NULL
	AND (? = 0 OR id != ?);

-- name: GetClassRecordDuplicate :one
SELECT
	cr.id,
	cr.date,
	cr.start_time,
	cr.end_time,
	cr.duration_minutes,
	cr.status,
	s.name as student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE cr.student_id = ? AND cr.teacher_id = ? AND cr.date = ? AND cr.duration_minutes = ?
	AND cr.deleted_at IS NULL
	AND (? = 0 OR cr.id != ?)
LIMIT 1;

-- name: CountClassRecordsFiltered :one
SELECT COUNT(*) as count
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
WHERE (? = 0 OR cr.teacher_id = ?) AND cr.date >= ? AND cr.date <= ?
	AND cr.deleted_at IS NULL
	AND (? = '' OR cr.status = ?)
	AND (? = '' OR s.name LIKE '%' || ? || '%');

-- name: GetClassRecordsFiltered :many
SELECT cr.id, cr.student_id, cr.teacher_id, cr.date, cr.start_time, cr.end_time, cr.duration_minutes, cr.rate, cr.currency, cr.status, cr.reason, cr.notes, cr.created_at, cr.updated_at, cr.recorded_by_role,
	s.name as student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) as teacher_name
FROM tbl_class_records cr
JOIN tbl_students s ON cr.student_id = s.id
JOIN tbl_teachers t ON cr.teacher_id = t.id
WHERE (? = 0 OR cr.teacher_id = ?) AND cr.date >= ? AND cr.date <= ?
	AND cr.deleted_at IS NULL
	AND (? = '' OR cr.status = ?)
	AND (? = '' OR s.name LIKE '%' || ? || '%')
ORDER BY CASE WHEN cr.date = date('now', 'localtime') THEN 0 ELSE 1 END, cr.date DESC, cr.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountClassesListFiltered :one
SELECT (
	SELECT COUNT(*) as count
	FROM tbl_class_records cr
	JOIN tbl_students s ON cr.student_id = s.id
	WHERE (? = 0 OR cr.teacher_id = ?) AND cr.date >= ? AND cr.date <= ?
		AND ((? = 'deleted' AND cr.deleted_at IS NOT NULL) OR (? != 'deleted' AND cr.deleted_at IS NULL))
		AND (? = '' OR ? != 'scheduled')
		AND (? = '' OR ? = 'deleted' OR cr.status = ?)
		AND (? = '' OR s.name LIKE '%' || ? || '%')
) + (
	SELECT COUNT(*) as count
	FROM tbl_scheduled_classes sc
	JOIN tbl_students s ON sc.student_id = s.id
	WHERE (? = 0 OR sc.teacher_id = ?) AND sc.scheduled_date >= ? AND sc.scheduled_date <= ?
		AND ((? = 'deleted' AND sc.deleted_at IS NOT NULL) OR (? != 'deleted' AND sc.deleted_at IS NULL AND sc.status = 'scheduled'))
		AND (? = '' OR ? = 'scheduled' OR ? = 'deleted')
		AND (? = '' OR s.name LIKE '%' || ? || '%')
) AS count;

-- name: GetClassesListFiltered :many
SELECT * FROM (
	SELECT
		cr.id,
	'record' AS source,
	cr.student_id,
	cr.teacher_id,
	cr.date AS date,
	cr.start_time,
	cr.end_time,
	cr.duration_minutes,
	cr.rate,
	cr.currency,
	CAST(CASE WHEN cr.deleted_at IS NOT NULL THEN 'deleted' ELSE cr.status END AS TEXT) AS status,
	cr.reason,
	cr.notes,
	cr.created_at,
	s.name AS student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) AS teacher_name,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.assigned_color AS teacher_assigned_color,
	t.profile_picture AS teacher_profile_picture
	FROM tbl_class_records cr
	JOIN tbl_students s ON cr.student_id = s.id
	JOIN tbl_teachers t ON cr.teacher_id = t.id
	WHERE (? = 0 OR cr.teacher_id = ?) AND cr.date >= ? AND cr.date <= ?
		AND ((? = 'deleted' AND cr.deleted_at IS NOT NULL) OR (? != 'deleted' AND cr.deleted_at IS NULL))
		AND (? = '' OR ? != 'scheduled')
		AND (? = '' OR ? = 'deleted' OR cr.status = ?)
		AND (? = '' OR s.name LIKE '%' || ? || '%')

	UNION ALL

	SELECT
		sc.id,
	'scheduled' AS source,
	sc.student_id,
	sc.teacher_id,
	sc.scheduled_date AS date,
	sc.start_time,
	NULL AS end_time,
	sc.duration_minutes,
	sc.rate,
	sc.currency,
	CAST(CASE WHEN sc.deleted_at IS NOT NULL THEN 'deleted' ELSE sc.status END AS TEXT) AS status,
	sc.reason,
	NULL AS notes,
	sc.created_at,
	s.name AS student_name,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) AS teacher_name,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.assigned_color AS teacher_assigned_color,
	t.profile_picture AS teacher_profile_picture
	FROM tbl_scheduled_classes sc
	JOIN tbl_students s ON sc.student_id = s.id
	JOIN tbl_teachers t ON sc.teacher_id = t.id
	WHERE (? = 0 OR sc.teacher_id = ?) AND sc.scheduled_date >= ? AND sc.scheduled_date <= ?
		AND ((? = 'deleted' AND sc.deleted_at IS NOT NULL) OR (? != 'deleted' AND sc.deleted_at IS NULL AND sc.status = 'scheduled'))
		AND (? = '' OR ? = 'scheduled' OR ? = 'deleted')
		AND (? = '' OR s.name LIKE '%' || ? || '%')
) AS combined
ORDER BY CASE WHEN combined.date = date('now', 'localtime') THEN 0 ELSE 1 END, combined.date DESC, combined.start_time DESC, combined.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountClassRecordsByStatusAndDateRange :many
SELECT cr.status, COUNT(*) as count
FROM tbl_class_records cr
WHERE cr.date >= ? AND cr.date <= ?
	AND cr.deleted_at IS NULL
	AND (? = 0 OR cr.teacher_id = ?)
GROUP BY cr.status;

-- name: SumConductedRateByCurrencyAndDateRange :many
SELECT cr.currency, COALESCE(SUM(cr.rate), 0) as total_rate
FROM tbl_class_records cr
WHERE cr.date >= ? AND cr.date <= ? AND cr.status = 'conducted'
	AND cr.deleted_at IS NULL
	AND (? = 0 OR cr.teacher_id = ?)
GROUP BY cr.currency;
