-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_class_records_unique_session;

UPDATE tbl_class_records
SET
	start_time = strftime('%H:%M', datetime(date || ' ' || trim(end_time), '-' || duration_minutes || ' minutes')),
	updated_at = datetime('now')
WHERE deleted_at IS NULL
	AND (start_time IS NULL OR trim(start_time) = '')
	AND end_time IS NOT NULL
	AND trim(end_time) != ''
	AND duration_minutes > 0;

UPDATE tbl_class_records
SET deleted_at = datetime('now'), updated_at = datetime('now')
WHERE deleted_at IS NULL
	AND id IN (
		SELECT cr.id
		FROM tbl_class_records cr
		INNER JOIN (
			SELECT
				student_id,
				teacher_id,
				date,
				COALESCE(trim(start_time), '') AS session_start,
				MIN(id) AS keep_id
			FROM tbl_class_records
			WHERE deleted_at IS NULL
			GROUP BY student_id, teacher_id, date, COALESCE(trim(start_time), '')
			HAVING COUNT(*) > 1
		) d ON cr.student_id = d.student_id
			AND cr.teacher_id = d.teacher_id
			AND cr.date = d.date
			AND COALESCE(trim(cr.start_time), '') = d.session_start
			AND cr.id != d.keep_id
		WHERE cr.deleted_at IS NULL
	);

CREATE UNIQUE INDEX idx_class_records_unique_session
	ON tbl_class_records (student_id, teacher_id, date, COALESCE(trim(start_time), ''))
	WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_scheduled_classes_unique_session;

UPDATE tbl_scheduled_classes
SET deleted_at = datetime('now'), updated_at = datetime('now')
WHERE deleted_at IS NULL
	AND status = 'scheduled'
	AND id IN (
		SELECT sc.id
		FROM tbl_scheduled_classes sc
		INNER JOIN (
			SELECT
				student_id,
				teacher_id,
				scheduled_date,
				COALESCE(trim(start_time), '') AS session_start,
				MIN(id) AS keep_id
			FROM tbl_scheduled_classes
			WHERE deleted_at IS NULL
				AND status = 'scheduled'
			GROUP BY student_id, teacher_id, scheduled_date, COALESCE(trim(start_time), '')
			HAVING COUNT(*) > 1
		) d ON sc.student_id = d.student_id
			AND sc.teacher_id = d.teacher_id
			AND sc.scheduled_date = d.scheduled_date
			AND COALESCE(trim(sc.start_time), '') = d.session_start
			AND sc.id != d.keep_id
		WHERE sc.deleted_at IS NULL
			AND sc.status = 'scheduled'
	);

CREATE UNIQUE INDEX idx_scheduled_classes_unique_session
	ON tbl_scheduled_classes (student_id, teacher_id, scheduled_date, COALESCE(trim(start_time), ''))
	WHERE status = 'scheduled' AND deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_scheduled_classes_unique_session;
CREATE UNIQUE INDEX idx_scheduled_classes_unique_session
	ON tbl_scheduled_classes (student_id, teacher_id, scheduled_date, duration_minutes)
	WHERE status = 'scheduled' AND deleted_at IS NULL;

DROP INDEX IF EXISTS idx_class_records_unique_session;
CREATE UNIQUE INDEX idx_class_records_unique_session
	ON tbl_class_records (student_id, teacher_id, date, duration_minutes)
	WHERE deleted_at IS NULL;
-- +goose StatementEnd
