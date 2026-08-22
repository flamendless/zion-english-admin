-- +goose Up
ALTER TABLE tbl_teachers DROP COLUMN name;

-- +goose Down
ALTER TABLE tbl_teachers ADD COLUMN name TEXT NOT NULL DEFAULT '';
UPDATE tbl_teachers SET name = trim(first_name || ' ' || middle_name || ' ' || last_name);
