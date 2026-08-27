-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_class_records ADD COLUMN notes TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_class_records_unique_session
	ON tbl_class_records (student_id, teacher_id, date, duration_minutes);

CREATE INDEX IF NOT EXISTS idx_students_status ON tbl_students(status);
CREATE INDEX IF NOT EXISTS idx_students_name ON tbl_students(name);
CREATE INDEX IF NOT EXISTS idx_teachers_status ON tbl_teachers(status);
CREATE INDEX IF NOT EXISTS idx_logs_module_created_at ON tbl_logs(module, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_class_records_teacher_date ON tbl_class_records(teacher_id, date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_class_records_teacher_date;
DROP INDEX IF EXISTS idx_logs_module_created_at;
DROP INDEX IF EXISTS idx_teachers_status;
DROP INDEX IF EXISTS idx_students_name;
DROP INDEX IF EXISTS idx_students_status;
DROP INDEX IF EXISTS idx_class_records_unique_session;
ALTER TABLE tbl_class_records DROP COLUMN notes;
-- +goose StatementEnd
