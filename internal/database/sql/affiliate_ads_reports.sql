-- name: GetAffiliateClickSummary :one
SELECT
	COUNT(*) AS product_count,
	COALESCE(SUM(click_count), 0) AS total_clicks,
	COUNT(CASE WHEN click_count > 0 THEN 1 END) AS products_with_clicks
FROM tbl_affiliate_products;

-- name: CountAffiliateProductsForClickReport :one
SELECT COUNT(*) AS count
FROM tbl_affiliate_products p
LEFT JOIN tbl_affiliated_product_shops s ON s.id = p.affiliated_shop_id
WHERE (
	? = ''
	OR p.name LIKE '%' || ? || '%'
	OR COALESCE(s.brand_name, p.brand, '') LIKE '%' || ? || '%'
);

-- name: ListAffiliateProductsForClickReport :many
SELECT
	p.id,
	p.name,
	p.thumbnail_url,
	p.click_count,
	COALESCE(s.brand_name, p.brand, '') AS shop_name
FROM tbl_affiliate_products p
LEFT JOIN tbl_affiliated_product_shops s ON s.id = p.affiliated_shop_id
WHERE (
	? = ''
	OR p.name LIKE '%' || ? || '%'
	OR COALESCE(s.brand_name, p.brand, '') LIKE '%' || ? || '%'
)
ORDER BY p.click_count DESC, p.id ASC
LIMIT ? OFFSET ?;

-- name: ListAdsClickReport :many
SELECT
	a.id,
	a.name,
	a.placement,
	a.status,
	a.sort_order,
	COUNT(j.affiliate_product_id) AS product_count,
	COALESCE(SUM(p.click_count), 0) AS linked_click_total
FROM tbl_ads a
LEFT JOIN tbl_ad_affiliate_products j ON j.ad_id = a.id
LEFT JOIN tbl_affiliate_products p ON p.id = j.affiliate_product_id
WHERE a.status != 'deleted'
GROUP BY a.id
ORDER BY a.sort_order ASC, a.id ASC;
