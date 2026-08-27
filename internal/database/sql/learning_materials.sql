-- name: GetTeacherNameByID :one
SELECT first_name, middle_name, last_name
FROM tbl_teachers
WHERE id = ? AND deleted = 0;

-- name: CountLearningMaterialTags :one
SELECT COUNT(*) FROM tbl_learning_material_tags;

-- name: GetLearningMaterialTagByLabel :one
SELECT id, label, color, created_at
FROM tbl_learning_material_tags
WHERE label = ? COLLATE NOCASE;

-- name: InsertLearningMaterialTag :one
INSERT INTO tbl_learning_material_tags (label, color)
VALUES (?, ?)
RETURNING id;

-- name: SearchLearningMaterialTags :many
SELECT id, label, color, created_at
FROM tbl_learning_material_tags
WHERE ? = '' OR label LIKE '%' || ? || '%'
ORDER BY label ASC
LIMIT 20;

-- name: GetAllLearningMaterialTags :many
SELECT id, label, color, created_at
FROM tbl_learning_material_tags
ORDER BY label ASC;

-- name: InsertLearningMaterial :one
INSERT INTO tbl_learning_materials (owner_id, title, description, url, thumbnail_url, access, status)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateLearningMaterial :exec
UPDATE tbl_learning_materials
SET
	title = sqlc.arg(title),
	description = sqlc.arg(description),
	url = sqlc.arg(url),
	thumbnail_url = sqlc.arg(thumbnail_url),
	access = sqlc.arg(access),
	status = sqlc.arg(status),
	deleted_at = CASE
		WHEN sqlc.arg(status) = 'deleted' AND deleted_at IS NULL THEN datetime('now')
		WHEN sqlc.arg(status) = 'draft' OR sqlc.arg(status) = 'published' THEN NULL
		ELSE deleted_at
	END,
	updated_at = datetime('now')
WHERE id = sqlc.arg(id);

-- name: ClearLearningMaterialDeleted :exec
UPDATE tbl_learning_materials
SET deleted_at = NULL, updated_at = datetime('now')
WHERE id = ? AND status != 'deleted' AND deleted_at IS NOT NULL;

-- name: DeleteLearningMaterial :exec
UPDATE tbl_learning_materials
SET status = 'deleted', deleted_at = datetime('now'), updated_at = datetime('now')
WHERE id = ? AND status != 'deleted';

-- name: GetLearningMaterialByID :one
SELECT id, owner_id, title, description, url, thumbnail_url, access, status, created_at, updated_at, deleted_at
FROM tbl_learning_materials
WHERE id = ?;

-- name: CountLearningMaterialsForSuperuser :one
SELECT COUNT(*) FROM tbl_learning_materials;

-- name: GetLearningMaterialsPagedForSuperuser :many
SELECT id, owner_id, title, description, url, thumbnail_url, access, status, created_at, updated_at, deleted_at
FROM tbl_learning_materials
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountLearningMaterialsForUser :one
SELECT COUNT(*)
FROM tbl_learning_materials m
WHERE m.owner_id = ?
	OR (
		m.status = 'published'
		AND m.access = 'public'
		AND m.status != 'deleted'
	);

-- name: GetLearningMaterialsPagedForUser :many
SELECT id, owner_id, title, description, url, thumbnail_url, access, status, created_at, updated_at, deleted_at
FROM tbl_learning_materials m
WHERE m.owner_id = ?
	OR (
		m.status = 'published'
		AND m.access = 'public'
		AND m.status != 'deleted'
	)
ORDER BY m.created_at DESC, m.id DESC
LIMIT ? OFFSET ?;

-- name: DeleteLearningMaterialTagLinks :exec
DELETE FROM tbl_learning_materials_tags_m2m WHERE material_id = ?;

-- name: InsertLearningMaterialTagLink :exec
INSERT INTO tbl_learning_materials_tags_m2m (material_id, tag_id) VALUES (?, ?);

-- name: GetTagIDsByMaterialID :many
SELECT tag_id FROM tbl_learning_materials_tags_m2m WHERE material_id = ?;

-- name: GetTagsByMaterialID :many
SELECT t.id, t.label, t.color, t.created_at
FROM tbl_learning_material_tags t
INNER JOIN tbl_learning_materials_tags_m2m m ON m.tag_id = t.id
WHERE m.material_id = ?
ORDER BY t.label ASC;

-- name: GetTagsByMaterialIDs :many
SELECT m.material_id, t.id, t.label, t.color, t.created_at
FROM tbl_learning_material_tags t
INNER JOIN tbl_learning_materials_tags_m2m m ON m.tag_id = t.id
WHERE m.material_id IN (sqlc.slice('material_ids'))
ORDER BY t.label ASC;
