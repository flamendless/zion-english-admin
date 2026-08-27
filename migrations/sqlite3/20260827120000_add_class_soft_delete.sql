-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_class_records ADD COLUMN deleted_at TEXT;
ALTER TABLE tbl_scheduled_classes ADD COLUMN deleted_at TEXT;

DROP INDEX IF EXISTS idx_class_records_unique_session;
CREATE UNIQUE INDEX idx_class_records_unique_session
    ON tbl_class_records (student_id, teacher_id, date, duration_minutes)
    WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_scheduled_classes_unique_session;
CREATE UNIQUE INDEX idx_scheduled_classes_unique_session
    ON tbl_scheduled_classes (student_id, teacher_id, scheduled_date, duration_minutes)
    WHERE status = 'scheduled' AND deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_scheduled_classes_unique_session;
CREATE UNIQUE INDEX idx_scheduled_classes_unique_session
    ON tbl_scheduled_classes (student_id, teacher_id, scheduled_date, duration_minutes)
    WHERE status = 'scheduled';

DROP INDEX IF EXISTS idx_class_records_unique_session;
CREATE UNIQUE INDEX idx_class_records_unique_session
    ON tbl_class_records (student_id, teacher_id, date, duration_minutes);

ALTER TABLE tbl_scheduled_classes DROP COLUMN deleted_at;
ALTER TABLE tbl_class_records DROP COLUMN deleted_at;
-- +goose StatementEnd
