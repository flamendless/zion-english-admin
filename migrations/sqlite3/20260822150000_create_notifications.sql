-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_teacher_id INTEGER NULL REFERENCES tbl_teachers(id),
    from_name TEXT NOT NULL,
    to_teacher_id INTEGER NULL REFERENCES tbl_teachers(id),
    to_name TEXT NOT NULL,
    message TEXT NOT NULL,
    kind TEXT NOT NULL,
    dedupe_key TEXT NULL UNIQUE,
    read INTEGER NOT NULL DEFAULT 0,
    read_at TEXT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON tbl_notifications(to_teacher_id, read, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_notifications_recipient;
DROP TABLE IF EXISTS tbl_notifications;
-- +goose StatementEnd
