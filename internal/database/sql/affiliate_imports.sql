-- name: InsertAffiliateImportBatch :one
INSERT INTO tbl_affiliate_import_batches (
	filename,
	product_count,
	file_size_bytes,
	csv_row_count,
	created_by,
	created_by_name
)
VALUES (?, 0, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateAffiliateImportBatchCount :exec
UPDATE tbl_affiliate_import_batches
SET product_count = ?
WHERE id = ?;

-- name: GetAffiliateImportBatchByID :one
SELECT
	id,
	filename,
	product_count,
	file_size_bytes,
	csv_row_count,
	created_at,
	created_by,
	created_by_name
FROM tbl_affiliate_import_batches
WHERE id = ?;

-- name: GetAffiliateProductsByImportBatchID :many
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
WHERE p.import_batch_id = ?
ORDER BY p.sort_order ASC, p.id ASC;
