-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_teachers ADD COLUMN drive_url TEXT NOT NULL;
ALTER TABLE tbl_teachers ADD COLUMN sex TEXT CHECK(sex IN ('M', 'F'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_teachers DROP COLUMN drive_url;
ALTER TABLE tbl_teachers DROP COLUMN sex;
-- +goose StatementEnd
