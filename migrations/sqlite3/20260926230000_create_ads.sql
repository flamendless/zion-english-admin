-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_ads (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	placement TEXT NOT NULL,
	ad_type TEXT NOT NULL DEFAULT 'affiliate',
	status TEXT NOT NULL DEFAULT 'draft' CHECK(status IN ('published', 'draft', 'deleted')),
	randomize_kind TEXT NOT NULL DEFAULT 'per_page',
	timer_interval TEXT NOT NULL DEFAULT '',
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tbl_ad_affiliate_products (
	ad_id INTEGER NOT NULL,
	affiliate_product_id INTEGER NOT NULL,
	sort_order INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (ad_id, affiliate_product_id),
	FOREIGN KEY (ad_id) REFERENCES tbl_ads(id) ON DELETE CASCADE,
	FOREIGN KEY (affiliate_product_id) REFERENCES tbl_affiliate_products(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tbl_ads_status ON tbl_ads(status);
CREATE INDEX IF NOT EXISTS idx_tbl_ads_sort_order ON tbl_ads(sort_order);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_ad_affiliate_products;
DROP TABLE IF EXISTS tbl_ads;
-- +goose StatementEnd
