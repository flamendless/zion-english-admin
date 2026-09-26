-- name: GetAllMetaTags :many
SELECT
	id,
	name,
	content,
	value,
	sort_order,
	created_at,
	updated_at
FROM tbl_meta_tags
ORDER BY sort_order ASC, id ASC;

-- name: GetMetaTagByID :one
SELECT
	id,
	name,
	content,
	value,
	sort_order,
	created_at,
	updated_at
FROM tbl_meta_tags
WHERE id = ?;

-- name: GetMetaTagByName :one
SELECT
	id,
	name,
	content,
	value,
	sort_order,
	created_at,
	updated_at
FROM tbl_meta_tags
WHERE name = ?;

-- name: InsertMetaTag :one
INSERT INTO tbl_meta_tags (name, content, value, sort_order)
VALUES (?, ?, ?, ?)
RETURNING id;

-- name: UpdateMetaTag :exec
UPDATE tbl_meta_tags
SET
	name = ?,
	content = ?,
	value = ?,
	sort_order = ?,
	updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteMetaTag :exec
DELETE FROM tbl_meta_tags
WHERE id = ?;
