-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_teachers ADD COLUMN first_name TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_teachers ADD COLUMN middle_name TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_teachers ADD COLUMN last_name TEXT NOT NULL DEFAULT '';
UPDATE tbl_teachers SET first_name = name WHERE first_name = '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_teachers DROP COLUMN last_name;
ALTER TABLE tbl_teachers DROP COLUMN middle_name;
ALTER TABLE tbl_teachers DROP COLUMN first_name;
-- +goose StatementEnd
