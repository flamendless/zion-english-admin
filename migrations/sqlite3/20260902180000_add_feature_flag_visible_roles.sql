-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_feature_flags ADD COLUMN visible_roles TEXT NOT NULL DEFAULT 'teacher,admin,developer';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_feature_flags DROP COLUMN visible_roles;
-- +goose StatementEnd
