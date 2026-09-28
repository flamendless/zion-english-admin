-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_affiliate_link_click_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	affiliate_product_id INTEGER NOT NULL,
	teacher_id INTEGER,
	ad_id INTEGER,
	ad_zone TEXT,
	clicked_at TEXT NOT NULL DEFAULT (datetime('now')),
	FOREIGN KEY (affiliate_product_id) REFERENCES tbl_affiliate_products(id) ON DELETE CASCADE,
	FOREIGN KEY (teacher_id) REFERENCES tbl_teachers(id) ON DELETE SET NULL,
	FOREIGN KEY (ad_id) REFERENCES tbl_ads(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_affiliate_click_events_teacher_clicked
	ON tbl_affiliate_link_click_events(teacher_id, clicked_at DESC);

CREATE INDEX IF NOT EXISTS idx_affiliate_click_events_ad_zone
	ON tbl_affiliate_link_click_events(ad_id, ad_zone);

CREATE INDEX IF NOT EXISTS idx_affiliate_click_events_product_clicked
	ON tbl_affiliate_link_click_events(affiliate_product_id, clicked_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_affiliate_link_click_events;
-- +goose StatementEnd
