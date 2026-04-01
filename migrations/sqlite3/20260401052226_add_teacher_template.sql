-- +goose Up

ALTER TABLE tbl_teachers ADD COLUMN template TEXT;

-- +goose Down

ALTER TABLE tbl_teachers DROP COLUMN template;
