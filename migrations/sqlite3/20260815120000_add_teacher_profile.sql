-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_teachers ADD COLUMN profile_picture TEXT;
ALTER TABLE tbl_teachers ADD COLUMN password_changed_at DATETIME;
ALTER TABLE tbl_teachers ADD COLUMN mobile_changed_at DATETIME;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_teachers DROP COLUMN profile_picture;
ALTER TABLE tbl_teachers DROP COLUMN password_changed_at;
ALTER TABLE tbl_teachers DROP COLUMN mobile_changed_at;
-- +goose StatementEnd
