-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_class_records_unique_session;
CREATE UNIQUE INDEX idx_class_records_unique_session
	ON tbl_class_records (student_id, teacher_id, date, COALESCE(trim(start_time), ''))
	WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_scheduled_classes_unique_session;
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
