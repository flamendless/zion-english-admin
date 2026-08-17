-- +goose Up
ALTER TABLE tbl_class_records ADD COLUMN start_time TEXT;

-- +goose Down
ALTER TABLE tbl_class_records DROP COLUMN start_time;
