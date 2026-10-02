-- +goose Up
-- Resume Impact provider migration when 20261002120000 applied ALTERs but failed on junction dedupe.
-- If provider column already exists and goose still shows 20261002120000 pending, mark that version
-- applied manually, then run goose so this migration completes dedupe and the unique index.

-- +goose StatementBegin
INSERT OR IGNORE INTO tbl_affiliated_product_shops (brand_name)
VALUES ('Impact');

CREATE TEMP TABLE affiliate_product_keeper AS
SELECT
	provider,
	item_id,
	MAX(id) AS keep_id
FROM tbl_affiliate_products
WHERE trim(item_id) != ''
GROUP BY provider, item_id;

UPDATE tbl_affiliate_products
SET click_count = (
	SELECT COALESCE(SUM(p2.click_count), 0)
	FROM tbl_affiliate_products p2
	WHERE p2.provider = tbl_affiliate_products.provider
		AND p2.item_id = tbl_affiliate_products.item_id
)
WHERE id IN (SELECT keep_id FROM affiliate_product_keeper);

DELETE FROM tbl_ad_affiliate_products
WHERE rowid IN (
	SELECT j.rowid
	FROM tbl_ad_affiliate_products j
	INNER JOIN tbl_affiliate_products p ON p.id = j.affiliate_product_id
	WHERE trim(p.item_id) != ''
		AND j.rowid NOT IN (
			SELECT MIN(j3.rowid)
			FROM tbl_ad_affiliate_products j3
			INNER JOIN tbl_affiliate_products p3 ON p3.id = j3.affiliate_product_id
			WHERE j3.ad_id = j.ad_id
				AND p3.provider = p.provider
				AND p3.item_id = p.item_id
		)
);

UPDATE tbl_ad_affiliate_products
SET affiliate_product_id = (
	SELECT k.keep_id
	FROM affiliate_product_keeper k
	INNER JOIN tbl_affiliate_products p ON p.id = tbl_ad_affiliate_products.affiliate_product_id
	WHERE k.provider = p.provider
		AND k.item_id = p.item_id
)
WHERE affiliate_product_id IN (
	SELECT p.id
	FROM tbl_affiliate_products p
	INNER JOIN affiliate_product_keeper k ON k.provider = p.provider AND k.item_id = p.item_id
	WHERE p.id != k.keep_id
);

UPDATE tbl_affiliate_link_click_events
SET affiliate_product_id = (
	SELECT k.keep_id
	FROM affiliate_product_keeper k
	INNER JOIN tbl_affiliate_products p ON p.id = tbl_affiliate_link_click_events.affiliate_product_id
	WHERE k.provider = p.provider
		AND k.item_id = p.item_id
)
WHERE affiliate_product_id IN (
	SELECT p.id
	FROM tbl_affiliate_products p
	INNER JOIN affiliate_product_keeper k ON k.provider = p.provider AND k.item_id = p.item_id
	WHERE p.id != k.keep_id
);

DELETE FROM tbl_ad_affiliate_products
WHERE rowid NOT IN (
	SELECT MIN(rowid)
	FROM tbl_ad_affiliate_products
	GROUP BY ad_id, affiliate_product_id
);

DELETE FROM tbl_affiliate_products
WHERE trim(item_id) != ''
	AND id NOT IN (SELECT keep_id FROM affiliate_product_keeper);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS idx_affiliate_products_provider_item_id
	ON tbl_affiliate_products (provider, item_id)
	WHERE trim(item_id) != '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_affiliate_products_provider_item_id;
-- +goose StatementEnd
