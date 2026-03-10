-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_teachers ADD COLUMN password TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- This is a no-op for SQLite as ALTER TABLE DROP COLUMN is not supported
-- +goose StatementEnd
