-- name: GetAllAds :many
SELECT
	id,
	name,
	placement,
	ad_type,
	status,
	randomize_kind,
	timer_interval,
	sort_order,
	created_at,
	updated_at
FROM tbl_ads
ORDER BY sort_order ASC, id ASC;

-- name: GetPublishedAdsWithProducts :many
SELECT
	a.id AS ad_id,
	a.name AS ad_name,
	a.placement,
	a.ad_type,
	a.randomize_kind,
	a.timer_interval,
	a.sort_order AS ad_sort_order,
	j.affiliate_product_id,
	j.sort_order AS product_sort_order,
	p.name AS product_name,
	p.price_display,
	p.thumbnail_url,
	p.affiliate_url,
	COALESCE(s.brand_name, p.brand, '') AS shop_name,
	p.sales,
	p.provider,
	p.program_id,
	p.item_id AS product_item_id
FROM tbl_ads a
INNER JOIN tbl_ad_affiliate_products j ON j.ad_id = a.id
INNER JOIN tbl_affiliate_products p ON p.id = j.affiliate_product_id
LEFT JOIN tbl_affiliated_product_shops s ON s.id = p.affiliated_shop_id
WHERE a.status = 'published'
	AND (
		p.provider != 'impact'
		OR upper(p.impact_state) = 'ACTIVE'
	)
ORDER BY a.sort_order ASC, a.id ASC, j.sort_order ASC, j.affiliate_product_id ASC;

-- name: GetAdByID :one
SELECT
	id,
	name,
	placement,
	ad_type,
	status,
	randomize_kind,
	timer_interval,
	sort_order,
	created_at,
	updated_at
FROM tbl_ads
WHERE id = ?;

-- name: InsertAd :one
INSERT INTO tbl_ads (
	name,
	placement,
	ad_type,
	status,
	randomize_kind,
	timer_interval,
	sort_order
) VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateAd :exec
UPDATE tbl_ads
SET
	name = ?,
	placement = ?,
	ad_type = ?,
	status = ?,
	randomize_kind = ?,
	timer_interval = ?,
	sort_order = ?,
	updated_at = datetime('now')
WHERE id = ?;

-- name: SoftDeleteAd :exec
UPDATE tbl_ads
SET status = 'deleted', updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteAdAffiliateProducts :exec
DELETE FROM tbl_ad_affiliate_products
WHERE ad_id = ?;

-- name: InsertAdAffiliateProduct :exec
INSERT INTO tbl_ad_affiliate_products (ad_id, affiliate_product_id, sort_order)
VALUES (?, ?, ?);

-- name: GetAdAffiliateProductIDs :many
SELECT affiliate_product_id
FROM tbl_ad_affiliate_products
WHERE ad_id = ?
ORDER BY sort_order ASC, affiliate_product_id ASC;

-- name: CountAdAffiliateProducts :one
SELECT COUNT(*) AS count
FROM tbl_ad_affiliate_products
WHERE ad_id = ?;

-- name: CountAffiliateProductsForAds :one
SELECT COUNT(*) AS count
FROM tbl_affiliate_products p
LEFT JOIN tbl_affiliated_product_shops s ON s.id = p.affiliated_shop_id
WHERE (
	? = ''
	OR p.name LIKE '%' || ? || '%'
	OR COALESCE(s.brand_name, p.brand, '') LIKE '%' || ? || '%'
)
	AND (
		p.provider != 'impact'
		OR upper(p.impact_state) = 'ACTIVE'
	);

-- name: SearchAffiliateProductsForAds :many
SELECT
	p.id,
	p.name,
	p.price_display,
	p.thumbnail_url,
	COALESCE(s.brand_name, p.brand, '') AS shop_name,
	p.provider,
	p.program_id,
	p.item_id
FROM tbl_affiliate_products p
LEFT JOIN tbl_affiliated_product_shops s ON s.id = p.affiliated_shop_id
WHERE (
	? = ''
	OR p.name LIKE '%' || ? || '%'
	OR COALESCE(s.brand_name, p.brand, '') LIKE '%' || ? || '%'
)
	AND (
		p.provider != 'impact'
		OR upper(p.impact_state) = 'ACTIVE'
	)
ORDER BY p.sort_order ASC, p.id ASC
LIMIT ? OFFSET ?;

-- name: ListAffiliateProductIDsForAds :many
SELECT p.id
FROM tbl_affiliate_products p
LEFT JOIN tbl_affiliated_product_shops s ON s.id = p.affiliated_shop_id
WHERE (
	? = ''
	OR p.name LIKE '%' || ? || '%'
	OR COALESCE(s.brand_name, p.brand, '') LIKE '%' || ? || '%'
)
	AND (
		p.provider != 'impact'
		OR upper(p.impact_state) = 'ACTIVE'
	)
ORDER BY p.sort_order ASC, p.id ASC;
