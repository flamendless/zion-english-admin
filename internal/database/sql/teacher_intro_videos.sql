-- name: InsertTeacherIntroVideo :exec
INSERT INTO tbl_teacher_intro_videos (
	teacher_id, original_filename, stored_filename, mime_type, file_size, status
) VALUES (?, ?, ?, ?, ?, ?);

-- name: GetTeacherIntroVideoByID :one
SELECT
	id,
	teacher_id,
	original_filename,
	stored_filename,
	mime_type,
	file_size,
	status,
	created_at,
	reviewed_at,
	reviewed_by,
	deleted_at,
	reject_reason
FROM tbl_teacher_intro_videos
WHERE id = ?;

-- name: GetLatestTeacherIntroVideoByTeacherID :one
SELECT
	id,
	teacher_id,
	original_filename,
	stored_filename,
	mime_type,
	file_size,
	status,
	created_at,
	reviewed_at,
	reviewed_by,
	deleted_at,
	reject_reason
FROM tbl_teacher_intro_videos
WHERE teacher_id = ?
	AND status != 'deleted'
ORDER BY created_at DESC
LIMIT 1;

-- name: HasBlockingTeacherIntroVideo :one
SELECT COUNT(*) AS count
FROM tbl_teacher_intro_videos
WHERE teacher_id = ?
	AND status IN ('submitted', 'approved');

-- name: UpdateTeacherIntroVideoStatus :exec
UPDATE tbl_teacher_intro_videos
SET status = ?,
	reviewed_at = CURRENT_TIMESTAMP,
	reviewed_by = ?,
	reject_reason = ?
WHERE id = ?;

-- name: SoftDeleteTeacherIntroVideo :exec
UPDATE tbl_teacher_intro_videos
SET status = 'deleted',
	deleted_at = CURRENT_TIMESTAMP,
	reviewed_at = CURRENT_TIMESTAMP,
	reviewed_by = ?
WHERE id = ?;

-- name: CountTeacherIntroVideosByStatus :one
SELECT COUNT(*) AS count
FROM tbl_teacher_intro_videos v
INNER JOIN tbl_teachers t ON t.id = v.teacher_id
WHERE v.status = ?
	AND t.deleted = 0;

-- name: GetAllTeacherIntroVideosFiltered :many
SELECT
	v.id,
	v.teacher_id,
	v.original_filename,
	v.stored_filename,
	v.mime_type,
	v.file_size,
	v.status,
	v.created_at,
	v.reviewed_at,
	v.reviewed_by,
	v.deleted_at,
	v.reject_reason,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) AS teacher_name,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.assigned_color AS teacher_assigned_color,
	t.profile_picture AS teacher_profile_picture
FROM tbl_teacher_intro_videos v
INNER JOIN tbl_teachers t ON t.id = v.teacher_id
WHERE t.deleted = 0
	AND (? = '' OR v.status = ?)
	AND (
	? = 0
	OR v.teacher_id = ?
	)
	AND (
	? = ''
	OR v.original_filename LIKE '%' || ? || '%'
	OR trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) LIKE '%' || ? || '%'
	)
ORDER BY v.created_at DESC;

-- name: GetTeacherIntroVideosByTeacherIDFiltered :many
SELECT
	id,
	teacher_id,
	original_filename,
	stored_filename,
	mime_type,
	file_size,
	status,
	created_at,
	reviewed_at,
	reviewed_by,
	deleted_at,
	reject_reason
FROM tbl_teacher_intro_videos
WHERE teacher_id = ?
	AND (? = '' OR status = ?)
	AND (
	? = ''
	OR original_filename LIKE '%' || ? || '%'
	)
ORDER BY created_at DESC;
