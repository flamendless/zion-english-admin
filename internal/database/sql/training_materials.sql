-- name: CountTrainingMaterialTags :one
SELECT COUNT(*) FROM tbl_training_material_tags;

-- name: GetTrainingMaterialTagByLabel :one
SELECT id, label, color, created_at
FROM tbl_training_material_tags
WHERE label = ? COLLATE NOCASE;

-- name: InsertTrainingMaterialTag :one
INSERT INTO tbl_training_material_tags (label, color)
VALUES (?, ?)
RETURNING id;

-- name: GetAllTrainingMaterialTags :many
SELECT id, label, color, created_at
FROM tbl_training_material_tags
ORDER BY label ASC;

-- name: InsertTrainingMaterial :one
INSERT INTO tbl_training_materials (
	title,
	description,
	url,
	embed_url,
	source_type,
	video_id,
	thumbnail_url,
	status,
	required,
	created_by
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateTrainingMaterial :exec
UPDATE tbl_training_materials
SET
	title = sqlc.arg(title),
	description = sqlc.arg(description),
	url = sqlc.arg(url),
	embed_url = sqlc.arg(embed_url),
	source_type = sqlc.arg(source_type),
	video_id = sqlc.arg(video_id),
	thumbnail_url = sqlc.arg(thumbnail_url),
	status = sqlc.arg(status),
	required = sqlc.arg(required),
	deleted_at = CASE
		WHEN sqlc.arg(status) = 'deleted' AND deleted_at IS NULL THEN datetime('now')
		WHEN sqlc.arg(status) = 'draft' OR sqlc.arg(status) = 'published' THEN NULL
		ELSE deleted_at
	END,
	updated_at = datetime('now')
WHERE id = sqlc.arg(id);

-- name: UpdateTrainingMaterialDuration :exec
UPDATE tbl_training_materials
SET
	duration_seconds = sqlc.arg(duration_seconds),
	updated_at = datetime('now')
WHERE id = sqlc.arg(id)
	AND (duration_seconds IS NULL OR duration_seconds = 0);

-- name: DeleteTrainingMaterial :exec
UPDATE tbl_training_materials
SET status = 'deleted', deleted_at = datetime('now'), updated_at = datetime('now')
WHERE id = ? AND status != 'deleted';

-- name: GetTrainingMaterialByID :one
SELECT
	id,
	title,
	description,
	url,
	embed_url,
	source_type,
	video_id,
	thumbnail_url,
	duration_seconds,
	status,
	required,
	created_by,
	created_at,
	updated_at,
	deleted_at
FROM tbl_training_materials
WHERE id = ?;

-- name: CountTrainingMaterialsForAdmin :one
SELECT COUNT(*) FROM tbl_training_materials;

-- name: GetTrainingMaterialsPagedForAdmin :many
SELECT
	id,
	title,
	description,
	url,
	embed_url,
	source_type,
	video_id,
	thumbnail_url,
	duration_seconds,
	status,
	required,
	created_by,
	created_at,
	updated_at,
	deleted_at
FROM tbl_training_materials
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: CountTrainingMaterialsPublished :one
SELECT COUNT(*)
FROM tbl_training_materials
WHERE status = 'published';

-- name: CountTrainingMaterialsRequiredPublished :one
SELECT COUNT(*)
FROM tbl_training_materials
WHERE status = 'published'
	AND required = 1;

-- name: GetTrainingMaterialsPagedPublished :many
SELECT
	id,
	title,
	description,
	url,
	embed_url,
	source_type,
	video_id,
	thumbnail_url,
	duration_seconds,
	status,
	required,
	created_by,
	created_at,
	updated_at,
	deleted_at
FROM tbl_training_materials
WHERE status = 'published'
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?;

-- name: DeleteTrainingMaterialTagLinks :exec
DELETE FROM tbl_training_materials_tags_m2m WHERE material_id = ?;

-- name: InsertTrainingMaterialTagLink :exec
INSERT INTO tbl_training_materials_tags_m2m (material_id, tag_id) VALUES (?, ?);

-- name: GetTagsByTrainingMaterialID :many
SELECT t.id, t.label, t.color, t.created_at
FROM tbl_training_material_tags t
INNER JOIN tbl_training_materials_tags_m2m m ON m.tag_id = t.id
WHERE m.material_id = ?
ORDER BY t.label ASC;

-- name: GetTagsByTrainingMaterialIDs :many
SELECT m.material_id, t.id, t.label, t.color, t.created_at
FROM tbl_training_material_tags t
INNER JOIN tbl_training_materials_tags_m2m m ON m.tag_id = t.id
WHERE m.material_id IN (sqlc.slice('material_ids'))
ORDER BY t.label ASC;

-- name: GetTrainingMaterialProgressByTeacherAndMaterial :one
SELECT
	material_id,
	teacher_id,
	watch_seconds,
	progress_percent,
	completed_at,
	first_viewed_at,
	last_viewed_at
FROM tbl_training_material_progress
WHERE material_id = ? AND teacher_id = ?;

-- name: UpsertTrainingMaterialProgress :exec
INSERT INTO tbl_training_material_progress (
	material_id,
	teacher_id,
	watch_seconds,
	progress_percent,
	completed_at,
	first_viewed_at,
	last_viewed_at
)
VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))
ON CONFLICT(material_id, teacher_id) DO UPDATE SET
	watch_seconds = excluded.watch_seconds,
	progress_percent = excluded.progress_percent,
	completed_at = CASE
		WHEN tbl_training_material_progress.completed_at IS NOT NULL THEN tbl_training_material_progress.completed_at
		WHEN excluded.completed_at IS NOT NULL THEN excluded.completed_at
		ELSE NULL
	END,
	last_viewed_at = datetime('now');

-- name: GetTrainingMaterialProgressByMaterialID :many
SELECT
	p.material_id,
	p.teacher_id,
	p.watch_seconds,
	p.progress_percent,
	p.completed_at,
	p.first_viewed_at,
	p.last_viewed_at,
	t.first_name,
	t.middle_name,
	t.last_name,
	t.assigned_color,
	t.profile_picture
FROM tbl_training_material_progress p
INNER JOIN tbl_teachers t ON t.id = p.teacher_id
WHERE p.material_id = ?
	AND t.deleted = 0
ORDER BY p.last_viewed_at DESC;

-- name: GetTrainingMaterialProgressReport :many
SELECT
	p.material_id,
	p.teacher_id,
	p.watch_seconds,
	p.progress_percent,
	p.completed_at,
	p.first_viewed_at,
	p.last_viewed_at,
	m.title AS material_title,
	t.first_name,
	t.middle_name,
	t.last_name,
	t.assigned_color,
	t.profile_picture
FROM tbl_training_material_progress p
INNER JOIN tbl_training_materials m ON m.id = p.material_id
INNER JOIN tbl_teachers t ON t.id = p.teacher_id
WHERE t.deleted = 0
	AND (? = 0 OR p.material_id = ?)
	AND (
		? = 'all'
		OR (? = 'completed' AND p.completed_at IS NOT NULL)
		OR (? = 'in_progress' AND p.completed_at IS NULL AND p.progress_percent > 0)
		OR (? = 'not_started' AND p.progress_percent = 0)
	)
ORDER BY p.last_viewed_at DESC, m.title ASC
LIMIT ? OFFSET ?;

-- name: CountTrainingMaterialProgressReport :one
SELECT COUNT(*)
FROM tbl_training_material_progress p
INNER JOIN tbl_training_materials m ON m.id = p.material_id
INNER JOIN tbl_teachers t ON t.id = p.teacher_id
WHERE t.deleted = 0
	AND (? = 0 OR p.material_id = ?)
	AND (
		? = 'all'
		OR (? = 'completed' AND p.completed_at IS NOT NULL)
		OR (? = 'in_progress' AND p.completed_at IS NULL AND p.progress_percent > 0)
		OR (? = 'not_started' AND p.progress_percent = 0)
	);

-- name: CountTeachersCompletedTrainingMaterial :one
SELECT COUNT(*)
FROM tbl_training_material_progress
WHERE material_id = ?
	AND completed_at IS NOT NULL;

-- name: CountPublishedTrainingMaterialsCompletedByTeacher :one
SELECT COUNT(DISTINCT p.material_id)
FROM tbl_training_material_progress p
INNER JOIN tbl_training_materials m ON m.id = p.material_id
WHERE p.teacher_id = ?
	AND m.status = 'published'
	AND p.completed_at IS NOT NULL;

-- name: CountRequiredPublishedTrainingMaterialsCompletedByTeacher :one
SELECT COUNT(DISTINCT p.material_id)
FROM tbl_training_material_progress p
INNER JOIN tbl_training_materials m ON m.id = p.material_id
WHERE p.teacher_id = ?
	AND m.status = 'published'
	AND m.required = 1
	AND p.completed_at IS NOT NULL;

-- name: GetTrainingMaterialProgressByTeacherID :many
SELECT
	material_id,
	teacher_id,
	watch_seconds,
	progress_percent,
	completed_at,
	first_viewed_at,
	last_viewed_at
FROM tbl_training_material_progress
WHERE teacher_id = ?
	AND material_id IN (sqlc.slice('material_ids'));

-- name: GetAllTrainingMaterialsForSelect :many
SELECT id, title
FROM tbl_training_materials
WHERE status != 'deleted'
ORDER BY title ASC;
