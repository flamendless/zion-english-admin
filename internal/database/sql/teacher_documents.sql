-- name: InsertTeacherDocument :exec
INSERT INTO tbl_teacher_documents (
	teacher_id, type, original_filename, stored_filename, file_extension, file_size, status
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetTeacherDocumentsByTeacherID :many
SELECT
	id,
	teacher_id,
	type,
	original_filename,
	stored_filename,
	file_extension,
	file_size,
	status,
	uploaded_at,
	reviewed_at,
	reviewed_by
FROM tbl_teacher_documents
WHERE teacher_id = ?
ORDER BY uploaded_at DESC;

-- name: GetTeacherDocumentsByTeacherIDFiltered :many
SELECT
	id,
	teacher_id,
	type,
	original_filename,
	stored_filename,
	file_extension,
	file_size,
	status,
	uploaded_at,
	reviewed_at,
	reviewed_by
FROM tbl_teacher_documents
WHERE teacher_id = ?
	AND (? = '' OR type = ?)
	AND (? = '' OR status = ?)
	AND (? = '' OR original_filename LIKE '%' || ? || '%')
ORDER BY uploaded_at DESC;

-- name: GetAllTeacherDocuments :many
SELECT
	d.id,
	d.teacher_id,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) AS teacher_name,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.profile_picture AS teacher_profile_picture,
	t.assigned_color AS teacher_assigned_color,
	d.type,
	d.original_filename,
	d.stored_filename,
	d.file_extension,
	d.file_size,
	d.status,
	d.uploaded_at,
	d.reviewed_at,
	d.reviewed_by
FROM tbl_teacher_documents d
INNER JOIN tbl_teachers t ON t.id = d.teacher_id
WHERE t.deleted = 0
ORDER BY d.uploaded_at DESC;

-- name: GetAllTeacherDocumentsFiltered :many
SELECT
	d.id,
	d.teacher_id,
	trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) AS teacher_name,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.profile_picture AS teacher_profile_picture,
	t.assigned_color AS teacher_assigned_color,
	d.type,
	d.original_filename,
	d.stored_filename,
	d.file_extension,
	d.file_size,
	d.status,
	d.uploaded_at,
	d.reviewed_at,
	d.reviewed_by
FROM tbl_teacher_documents d
INNER JOIN tbl_teachers t ON t.id = d.teacher_id
WHERE t.deleted = 0
	AND (? = '' OR d.type = ?)
	AND (? = '' OR d.status = ?)
	AND (? = 0 OR d.teacher_id = ?)
	AND (
	? = ''
	OR d.original_filename LIKE '%' || ? || '%'
	OR trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END) LIKE '%' || ? || '%'
	)
ORDER BY d.uploaded_at DESC;

-- name: GetTeacherDocumentByID :one
SELECT
	id,
	teacher_id,
	type,
	original_filename,
	stored_filename,
	file_extension,
	file_size,
	status,
	uploaded_at,
	reviewed_at,
	reviewed_by
FROM tbl_teacher_documents
WHERE id = ?;

-- name: HasBlockingTeacherDocument :one
SELECT COUNT(*) AS count
FROM tbl_teacher_documents
WHERE teacher_id = ?
	AND type = 'document'
	AND status IN ('submitted', 'approved');

-- name: UpdateTeacherDocumentStatus :exec
UPDATE tbl_teacher_documents
SET status = ?,
	reviewed_at = CURRENT_TIMESTAMP,
	reviewed_by = ?
WHERE id = ?;

-- name: CountTeacherDocumentsByStatus :one
SELECT COUNT(*) AS count
FROM tbl_teacher_documents d
INNER JOIN tbl_teachers t ON t.id = d.teacher_id
WHERE t.deleted = 0
	AND d.status = ?;

-- name: DeleteTeacherDocument :exec
DELETE FROM tbl_teacher_documents WHERE id = ?;
