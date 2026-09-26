-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_affiliate_products (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	affiliate_url TEXT NOT NULL,
	product_url TEXT NOT NULL DEFAULT '',
	shop_id TEXT NOT NULL DEFAULT '',
	item_id TEXT NOT NULL DEFAULT '',
	name TEXT NOT NULL,
	brand TEXT NOT NULL DEFAULT '',
	price_display TEXT NOT NULL DEFAULT '',
	thumbnail_url TEXT NOT NULL DEFAULT '',
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_affiliate_products;
-- +goose StatementEnd
