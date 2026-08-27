-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_teacher_meeting_accounts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	service TEXT NOT NULL,
	external_user_id TEXT NOT NULL DEFAULT '',
	access_token TEXT NOT NULL,
	refresh_token TEXT NOT NULL,
	token_expires_at TEXT NULL,
	connected_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	UNIQUE(teacher_id, service)
);

CREATE INDEX IF NOT EXISTS idx_teacher_meeting_accounts_teacher_service
	ON tbl_teacher_meeting_accounts(teacher_id, service);

CREATE TABLE IF NOT EXISTS tbl_class_meeting_room (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	class_id INTEGER NOT NULL REFERENCES tbl_scheduled_classes(id),
	service TEXT NOT NULL,
	room_id TEXT NOT NULL,
	room_url TEXT NOT NULL,
	room_passcode TEXT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	deleted_at TEXT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_class_meeting_room_active_class
	ON tbl_class_meeting_room(class_id)
	WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_class_meeting_room_class_id
	ON tbl_class_meeting_room(class_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_class_meeting_room_class_id;
DROP INDEX IF EXISTS idx_class_meeting_room_active_class;
DROP TABLE IF EXISTS tbl_class_meeting_room;
DROP INDEX IF EXISTS idx_teacher_meeting_accounts_teacher_service;
DROP TABLE IF EXISTS tbl_teacher_meeting_accounts;
-- +goose StatementEnd
