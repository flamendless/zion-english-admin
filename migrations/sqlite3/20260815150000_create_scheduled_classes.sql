-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_scheduled_classes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	student_id INTEGER NOT NULL REFERENCES tbl_students(id),
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	scheduled_date TEXT NOT NULL,
	start_time TEXT,
	duration_minutes INTEGER NOT NULL,
	rate REAL NOT NULL,
	currency TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'scheduled'
		CHECK(status IN ('scheduled', 'conducted', 'cancelled', 'rescheduled')),
	reason TEXT,
	created_by_role TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_scheduled_classes_teacher_date ON tbl_scheduled_classes(teacher_id, scheduled_date);
CREATE INDEX IF NOT EXISTS idx_scheduled_classes_date_status ON tbl_scheduled_classes(scheduled_date, status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_scheduled_classes_unique_session
	ON tbl_scheduled_classes (student_id, teacher_id, scheduled_date, duration_minutes)
	WHERE status = 'scheduled';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_scheduled_classes_unique_session;
DROP INDEX IF EXISTS idx_scheduled_classes_date_status;
DROP INDEX IF EXISTS idx_scheduled_classes_teacher_date;
DROP TABLE IF EXISTS tbl_scheduled_classes;
-- +goose StatementEnd
