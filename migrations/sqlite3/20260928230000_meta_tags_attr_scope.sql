-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_meta_tags ADD COLUMN attr TEXT NOT NULL DEFAULT 'name';
ALTER TABLE tbl_meta_tags ADD COLUMN scope TEXT NOT NULL DEFAULT 'site_wide';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite does not support DROP COLUMN in older versions; columns remain on rollback.
-- +goose StatementEnd
