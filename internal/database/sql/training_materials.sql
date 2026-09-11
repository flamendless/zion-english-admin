-- name: InsertTrainingMaterial :one
INSERT INTO tbl_training_materials (title, description, url, status, created_by_id)
VALUES (?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateTrainingMaterial :exec
UPDATE tbl_training_materials
SET title = ?, description = ?, url = ?, status = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteTrainingMaterial :exec
UPDATE tbl_training_materials
SET status = 'deleted', deleted_at = datetime('now'), updated_at = datetime('now')
WHERE id = ? AND status != 'deleted';

-- name: GetTrainingMaterialByID :one
SELECT id, title, description, url, status, created_by_id, created_at, updated_at, deleted_at
FROM tbl_training_materials
WHERE id = ?;

-- name: CountTrainingMaterialsForAdmin :one
SELECT COUNT(*) FROM tbl_training_materials;

-- name: GetTrainingMaterialsPagedForAdmin :many
SELECT id, title, description, url, status, created_by_id, created_at, updated_at, deleted_at
FROM tbl_training_materials
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountTrainingMaterialsForTeacher :one
SELECT COUNT(*)
FROM tbl_training_materials
WHERE status = 'published';

-- name: GetTrainingMaterialsPagedForTeacher :many
SELECT id, title, description, url, status, created_by_id, created_at, updated_at, deleted_at
FROM tbl_training_materials
WHERE status = 'published'
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;
