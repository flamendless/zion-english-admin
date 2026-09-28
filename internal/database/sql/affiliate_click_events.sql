-- name: InsertAffiliateLinkClickEvent :exec
INSERT INTO tbl_affiliate_link_click_events (
	affiliate_product_id,
	teacher_id,
	ad_id,
	ad_zone
) VALUES (?, ?, ?, ?);

-- name: CountAffiliateLinkClickEventsWithAd :one
SELECT COUNT(*) AS count
FROM tbl_affiliate_link_click_events
WHERE ad_id IS NOT NULL;

-- name: ListTopTeachersByAffiliateClickEvents :many
SELECT
	e.teacher_id,
	t.first_name,
	t.middle_name,
	t.last_name,
	t.assigned_color,
	t.profile_picture,
	COUNT(*) AS click_count
FROM tbl_affiliate_link_click_events e
INNER JOIN tbl_teachers t ON t.id = e.teacher_id
WHERE e.teacher_id IS NOT NULL
	AND t.deleted = 0
GROUP BY e.teacher_id
ORDER BY click_count DESC, e.teacher_id ASC
LIMIT ?;

-- name: ListAdZoneClickCounts :many
SELECT
	e.ad_id,
	a.name AS ad_name,
	e.ad_zone,
	COUNT(*) AS click_count
FROM tbl_affiliate_link_click_events e
INNER JOIN tbl_ads a ON a.id = e.ad_id
WHERE e.ad_id IS NOT NULL
	AND e.ad_zone IS NOT NULL
	AND e.ad_zone != ''
	AND a.status != 'deleted'
GROUP BY e.ad_id, e.ad_zone
ORDER BY click_count DESC, e.ad_id ASC, e.ad_zone ASC
LIMIT ?;

-- name: AdAffiliateProductInPublishedAd :one
SELECT 1 AS ok
FROM tbl_ads a
INNER JOIN tbl_ad_affiliate_products j ON j.ad_id = a.id
WHERE a.id = ?
	AND a.status = 'published'
	AND j.affiliate_product_id = ?
LIMIT 1;
