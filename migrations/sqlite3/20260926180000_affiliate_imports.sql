-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_affiliate_import_batches (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	filename TEXT NOT NULL DEFAULT '',
	product_count INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	created_by INTEGER,
	created_by_name TEXT NOT NULL DEFAULT ''
);

ALTER TABLE tbl_affiliate_products ADD COLUMN import_batch_id INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tbl_affiliate_products ADD COLUMN sales TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_affiliate_products ADD COLUMN shop_name TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_affiliate_products ADD COLUMN commission_rate TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_affiliate_products ADD COLUMN commission TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_affiliate_import_batches;
-- +goose StatementEnd
