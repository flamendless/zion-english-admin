-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_teacher_notification_preferences (
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	category TEXT NOT NULL,
	enabled INTEGER NOT NULL DEFAULT 1,
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (teacher_id, category)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_teacher_notification_preferences;
-- +goose StatementEnd
