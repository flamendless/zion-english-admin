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

-- name: GetAllTeacherDocuments :many
SELECT
    d.id,
    d.teacher_id,
    t.name AS teacher_name,
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
