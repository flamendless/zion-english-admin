-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_affiliate_import_batches ADD COLUMN file_size_bytes INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tbl_affiliate_import_batches ADD COLUMN csv_row_count INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- SQLite does not support DROP COLUMN in older versions; batch meta columns are left in place on rollback.
-- +goose StatementEnd
