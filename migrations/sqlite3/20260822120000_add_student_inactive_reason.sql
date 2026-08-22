-- +goose Up
ALTER TABLE tbl_students ADD COLUMN inactive_reason TEXT;

-- +goose Down
ALTER TABLE tbl_students DROP COLUMN inactive_reason;
