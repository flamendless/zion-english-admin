-- +goose Up
ALTER TABLE tbl_logs ADD COLUMN created_by_name TEXT;

UPDATE tbl_logs
SET created_by = NULL,
	created_by_name = 'superuser'
WHERE created_by IS NULL OR created_by = 0;

-- +goose Down
ALTER TABLE tbl_logs DROP COLUMN created_by_name;
