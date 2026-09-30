-- +goose Up
ALTER TABLE tbl_teachers ADD COLUMN resigned_at TEXT;
ALTER TABLE tbl_teachers ADD COLUMN resigned_reason TEXT;

-- +goose Down
ALTER TABLE tbl_teachers DROP COLUMN resigned_reason;
ALTER TABLE tbl_teachers DROP COLUMN resigned_at;
