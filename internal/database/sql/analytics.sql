-- name: GetAnalyticsSummary :one
SELECT
    COALESCE((
        SELECT SUM(CASE WHEN cr.status = 'conducted' THEN 1 ELSE 0 END)
        FROM tbl_class_records cr
        WHERE cr.date >= ? AND cr.date <= ?
          AND cr.deleted_at IS NULL
          AND (? = 0 OR cr.teacher_id = ?)
    ), 0) AS conducted,
    COALESCE((
        SELECT SUM(CASE WHEN cr.status = 'cancelled' THEN 1 ELSE 0 END)
        FROM tbl_class_records cr
        WHERE cr.date >= ? AND cr.date <= ?
          AND cr.deleted_at IS NULL
          AND (? = 0 OR cr.teacher_id = ?)
    ), 0) AS cancelled,
    COALESCE((
        SELECT SUM(CASE WHEN cr.status = 'rescheduled' THEN 1 ELSE 0 END)
        FROM tbl_class_records cr
        WHERE cr.date >= ? AND cr.date <= ?
          AND cr.deleted_at IS NULL
          AND (? = 0 OR cr.teacher_id = ?)
    ), 0) AS rescheduled,
    COALESCE((
        SELECT SUM(sc.duration_minutes)
        FROM tbl_scheduled_classes sc
        WHERE sc.scheduled_date >= ? AND sc.scheduled_date <= ?
          AND sc.deleted_at IS NULL
          AND (? = 0 OR sc.teacher_id = ?)
    ), 0) AS scheduled_minutes,
    COALESCE((
        SELECT SUM(cr.duration_minutes)
        FROM tbl_class_records cr
        WHERE cr.date >= ? AND cr.date <= ?
          AND cr.status = 'conducted'
          AND cr.deleted_at IS NULL
          AND (? = 0 OR cr.teacher_id = ?)
    ), 0) AS conducted_minutes,
    COALESCE((
        SELECT COUNT(*)
        FROM tbl_scheduled_classes sc
        WHERE sc.status = 'scheduled'
          AND sc.deleted_at IS NULL
          AND sc.scheduled_date < date('now', 'localtime')
          AND sc.scheduled_date >= ? AND sc.scheduled_date <= ?
          AND (? = 0 OR sc.teacher_id = ?)
    ), 0) AS no_show_count;

-- name: GetAnalyticsCancellationByTeacher :many
SELECT
    t.id AS teacher_id,
    trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) AS teacher_name,
    COALESCE(SUM(CASE WHEN cr.status = 'conducted' THEN 1 ELSE 0 END), 0) AS conducted,
    COALESCE(SUM(CASE WHEN cr.status = 'cancelled' THEN 1 ELSE 0 END), 0) AS cancelled,
    COALESCE(SUM(CASE WHEN cr.status = 'rescheduled' THEN 1 ELSE 0 END), 0) AS rescheduled,
    COALESCE((
        SELECT SUM(sc.duration_minutes)
        FROM tbl_scheduled_classes sc
        WHERE sc.teacher_id = t.id
          AND sc.scheduled_date >= ? AND sc.scheduled_date <= ?
          AND sc.deleted_at IS NULL
    ), 0) AS scheduled_minutes,
    COALESCE((
        SELECT SUM(cr2.duration_minutes)
        FROM tbl_class_records cr2
        WHERE cr2.teacher_id = t.id
          AND cr2.date >= ? AND cr2.date <= ?
          AND cr2.deleted_at IS NULL
          AND cr2.status = 'conducted'
    ), 0) AS conducted_minutes,
    COALESCE((
        SELECT COUNT(*)
        FROM tbl_scheduled_classes sc2
        WHERE sc2.teacher_id = t.id
          AND sc2.status = 'scheduled'
          AND sc2.deleted_at IS NULL
          AND sc2.scheduled_date < date('now', 'localtime')
          AND sc2.scheduled_date >= ? AND sc2.scheduled_date <= ?
    ), 0) AS no_show_count
FROM tbl_teachers t
LEFT JOIN tbl_class_records cr ON cr.teacher_id = t.id
    AND cr.date >= ? AND cr.date <= ?
    AND cr.deleted_at IS NULL
WHERE t.status = 'approved' AND t.deleted = 0
  AND (? = 0 OR t.id = ?)
GROUP BY t.id, t.first_name, t.middle_name, t.last_name
HAVING conducted + cancelled + rescheduled > 0
    OR scheduled_minutes > 0
    OR no_show_count > 0
ORDER BY cancelled DESC, teacher_name ASC;

-- name: GetAnalyticsCancellationByStudent :many
SELECT
    s.id AS student_id,
    s.name AS student_name,
    COALESCE(SUM(CASE WHEN cr.status = 'conducted' THEN 1 ELSE 0 END), 0) AS conducted,
    COALESCE(SUM(CASE WHEN cr.status = 'cancelled' THEN 1 ELSE 0 END), 0) AS cancelled,
    COALESCE(SUM(CASE WHEN cr.status = 'rescheduled' THEN 1 ELSE 0 END), 0) AS rescheduled
FROM tbl_students s
LEFT JOIN tbl_class_records cr ON cr.student_id = s.id
    AND cr.date >= ? AND cr.date <= ?
    AND cr.deleted_at IS NULL
    AND (? = 0 OR cr.teacher_id = ?)
WHERE (? = 0 OR EXISTS (
    SELECT 1 FROM tbl_teachers_students_m2m m
    WHERE m.student_id = s.id AND m.teacher_id = ?
))
GROUP BY s.id, s.name
HAVING conducted + cancelled + rescheduled > 0
ORDER BY cancelled DESC, student_name ASC;

-- name: GetAnalyticsWeeklyTrend :many
SELECT
    strftime('%Y-W%W', cr.date) AS week_label,
    COALESCE(SUM(CASE WHEN cr.status = 'conducted' THEN 1 ELSE 0 END), 0) AS conducted,
    COALESCE(SUM(CASE WHEN cr.status = 'cancelled' THEN 1 ELSE 0 END), 0) AS cancelled,
    COALESCE(SUM(CASE WHEN cr.status = 'rescheduled' THEN 1 ELSE 0 END), 0) AS rescheduled
FROM tbl_class_records cr
WHERE cr.date >= ? AND cr.date <= ?
  AND cr.deleted_at IS NULL
  AND (? = 0 OR cr.teacher_id = ?)
GROUP BY week_label
ORDER BY week_label ASC;

-- name: GetAnalyticsNoShows :many
SELECT
    sc.id,
    sc.scheduled_date,
    sc.start_time,
    sc.duration_minutes,
    s.name AS student_name,
    trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) AS teacher_name,
    CAST(julianday(date('now', 'localtime')) - julianday(sc.scheduled_date) AS INTEGER) AS days_overdue
FROM tbl_scheduled_classes sc
JOIN tbl_students s ON sc.student_id = s.id
JOIN tbl_teachers t ON sc.teacher_id = t.id
WHERE sc.status = 'scheduled'
  AND sc.deleted_at IS NULL
  AND sc.scheduled_date < date('now', 'localtime')
  AND sc.scheduled_date >= ? AND sc.scheduled_date <= ?
  AND (? = 0 OR sc.teacher_id = ?)
ORDER BY sc.scheduled_date ASC, sc.start_time ASC
LIMIT 50;

-- name: GetAnalyticsInactiveReasons :many
SELECT
    COALESCE(NULLIF(trim(s.inactive_reason), ''), '(not specified)') AS inactive_reason,
    COUNT(*) AS count
FROM tbl_students s
WHERE s.status = 'inactive'
  AND (? = 0 OR EXISTS (
    SELECT 1 FROM tbl_teachers_students_m2m m
    WHERE m.student_id = s.id AND m.teacher_id = ?
))
GROUP BY inactive_reason
ORDER BY count DESC, inactive_reason ASC;

-- name: GetAnalyticsChurnedStudents :many
SELECT
    s.id AS student_id,
    s.name AS student_name,
    COALESCE(s.inactive_reason, '') AS inactive_reason,
    CAST(julianday(s.updated_at) - julianday(s.created_at) AS INTEGER) AS tenure_days,
    s.updated_at AS churned_at
FROM tbl_students s
WHERE s.status = 'inactive'
  AND substr(s.updated_at, 1, 10) >= ? AND substr(s.updated_at, 1, 10) <= ?
  AND (? = 0 OR EXISTS (
    SELECT 1 FROM tbl_teachers_students_m2m m
    WHERE m.student_id = s.id AND m.teacher_id = ?
))
ORDER BY s.updated_at DESC, s.name ASC;

-- name: GetAnalyticsRetentionSummary :one
SELECT
    COALESCE((
        SELECT COUNT(*)
        FROM tbl_students s
        WHERE s.status = 'active'
          AND (? = 0 OR EXISTS (
            SELECT 1 FROM tbl_teachers_students_m2m m
            WHERE m.student_id = s.id AND m.teacher_id = ?
        ))
    ), 0) AS active_count,
    COALESCE((
        SELECT COUNT(*)
        FROM tbl_students s
        WHERE s.status = 'inactive'
          AND (? = 0 OR EXISTS (
            SELECT 1 FROM tbl_teachers_students_m2m m
            WHERE m.student_id = s.id AND m.teacher_id = ?
        ))
    ), 0) AS inactive_count,
    COALESCE((
        SELECT COUNT(*)
        FROM tbl_students s
        WHERE s.status = 'inactive'
          AND substr(s.updated_at, 1, 10) >= ? AND substr(s.updated_at, 1, 10) <= ?
          AND (? = 0 OR EXISTS (
            SELECT 1 FROM tbl_teachers_students_m2m m
            WHERE m.student_id = s.id AND m.teacher_id = ?
        ))
    ), 0) AS churned_in_period;
