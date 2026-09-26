-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_affiliated_product_shops (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	brand_name TEXT NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE tbl_affiliate_products ADD COLUMN affiliated_shop_id INTEGER REFERENCES tbl_affiliated_product_shops(id);
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO tbl_affiliated_product_shops (brand_name)
SELECT DISTINCT trim(shop_name)
FROM tbl_affiliate_products
WHERE trim(shop_name) != '';
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE tbl_affiliate_products
SET affiliated_shop_id = (
	SELECT s.id
	FROM tbl_affiliated_product_shops s
	WHERE s.brand_name = trim(tbl_affiliate_products.shop_name)
)
WHERE trim(shop_name) != '';
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE tbl_affiliate_products DROP COLUMN shop_name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_affiliate_products ADD COLUMN shop_name TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose StatementBegin
UPDATE tbl_affiliate_products
SET shop_name = COALESCE((
	SELECT s.brand_name
	FROM tbl_affiliated_product_shops s
	WHERE s.id = tbl_affiliate_products.affiliated_shop_id
), '');
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE tbl_affiliate_products DROP COLUMN affiliated_shop_id;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_affiliated_product_shops;
-- +goose StatementEnd
