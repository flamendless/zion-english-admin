-- +goose Up
ALTER TABLE tbl_class_records ADD COLUMN end_time TEXT;

-- +goose Down
ALTER TABLE tbl_class_records DROP COLUMN end_time;
