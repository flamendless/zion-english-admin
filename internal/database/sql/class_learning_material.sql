-- name: SearchLearningMaterialsByTitle :many
SELECT m.id, m.title, m.url
FROM tbl_learning_materials m
WHERE m.deleted_at IS NULL
	AND m.status != 'deleted'
	AND (? = '' OR m.title LIKE '%' || ? || '%')
	AND (
		? = 1
		OR m.owner_id = ?
		OR (m.status = 'published' AND m.access = 'public')
	)
ORDER BY m.title ASC
LIMIT 10;

-- name: GetLearningMaterialsByClassRecordID :many
SELECT m.id, m.title, m.url
FROM tbl_learning_materials m
INNER JOIN tbl_class_record_learning_materials l ON l.material_id = m.id
WHERE l.class_record_id = ?
	AND m.deleted_at IS NULL
	AND m.status != 'deleted'
ORDER BY m.title ASC;

-- name: DeleteClassRecordLearningMaterialLinks :exec
DELETE FROM tbl_class_record_learning_materials WHERE class_record_id = ?;

-- name: InsertClassRecordLearningMaterialLink :exec
INSERT INTO tbl_class_record_learning_materials (class_record_id, material_id)
VALUES (?, ?);

-- name: GetLearningMaterialsByScheduledClassID :many
SELECT m.id, m.title, m.url
FROM tbl_learning_materials m
INNER JOIN tbl_scheduled_class_learning_materials l ON l.material_id = m.id
WHERE l.scheduled_class_id = ?
	AND m.deleted_at IS NULL
	AND m.status != 'deleted'
ORDER BY m.title ASC;

-- name: DeleteScheduledClassLearningMaterialLinks :exec
DELETE FROM tbl_scheduled_class_learning_materials WHERE scheduled_class_id = ?;

-- name: InsertScheduledClassLearningMaterialLink :exec
INSERT INTO tbl_scheduled_class_learning_materials (scheduled_class_id, material_id)
VALUES (?, ?);
