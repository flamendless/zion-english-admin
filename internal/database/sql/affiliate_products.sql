-- name: GetAllAffiliateProducts :many
SELECT
	p.id,
	p.affiliate_url,
	p.product_url,
	p.shop_id,
	p.item_id,
	p.name,
	p.brand,
	p.price_display,
	p.thumbnail_url,
	p.sort_order,
	p.import_batch_id,
	p.sales,
	p.affiliated_shop_id,
	COALESCE(s.brand_name, '') AS shop_brand_name,
	p.commission_rate,
	p.commission,
	p.click_count,
	p.created_at,
	p.updated_at
FROM tbl_affiliate_products p
LEFT JOIN tbl_affiliated_product_shops s ON s.id = p.affiliated_shop_id
ORDER BY p.sort_order ASC, p.id ASC;

-- name: GetAffiliateProductByID :one
SELECT
	p.id,
	p.affiliate_url,
	p.product_url,
	p.shop_id,
	p.item_id,
	p.name,
	p.brand,
	p.price_display,
	p.thumbnail_url,
	p.sort_order,
	p.import_batch_id,
	p.sales,
	p.affiliated_shop_id,
	COALESCE(s.brand_name, '') AS shop_brand_name,
	p.commission_rate,
	p.commission,
	p.click_count,
	p.created_at,
	p.updated_at
FROM tbl_affiliate_products p
LEFT JOIN tbl_affiliated_product_shops s ON s.id = p.affiliated_shop_id
WHERE p.id = ?;

-- name: InsertAffiliateProduct :one
INSERT INTO tbl_affiliate_products (
	affiliate_url,
	product_url,
	shop_id,
	item_id,
	name,
	brand,
	price_display,
	thumbnail_url,
	sort_order,
	import_batch_id,
	sales,
	affiliated_shop_id,
	commission_rate,
	commission
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateAffiliateProduct :exec
UPDATE tbl_affiliate_products
SET
	affiliate_url = ?,
	product_url = ?,
	shop_id = ?,
	item_id = ?,
	name = ?,
	brand = ?,
	price_display = ?,
	thumbnail_url = ?,
	sort_order = ?,
	sales = ?,
	affiliated_shop_id = ?,
	commission_rate = ?,
	commission = ?,
	updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteAffiliateProduct :exec
DELETE FROM tbl_affiliate_products
WHERE id = ?;

-- name: IncrementAffiliateProductClickCount :exec
UPDATE tbl_affiliate_products
SET
	click_count = click_count + 1,
	updated_at = datetime('now')
WHERE id = ?;
