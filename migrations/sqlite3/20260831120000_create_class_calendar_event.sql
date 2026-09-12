-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_class_calendar_event (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	class_id INTEGER NOT NULL REFERENCES tbl_scheduled_classes(id),
	service TEXT NOT NULL DEFAULT 'google_calendar',
	event_id TEXT NOT NULL,
	event_url TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	deleted_at TEXT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_class_calendar_event_active
	ON tbl_class_calendar_event(class_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_class_calendar_event_active;
DROP TABLE IF EXISTS tbl_class_calendar_event;
