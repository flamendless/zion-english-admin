-- +goose Up
ALTER TABLE tbl_teacher_meeting_accounts ADD COLUMN resource_id TEXT NOT NULL DEFAULT '';

-- +goose Down
-- SQLite does not support DROP COLUMN in older versions; recreate would be required for full rollback.
