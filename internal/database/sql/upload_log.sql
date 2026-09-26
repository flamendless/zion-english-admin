-- name: InsertUploadLog :exec
INSERT INTO tbl_upload_logs (
	module,
	outcome,
	kind,
	summary,
	filename,
	file_size,
	compress_preset,
	created_by,
	created_by_name,
	created_at
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now')
);

-- name: CountUploadLogsFiltered :one
SELECT COUNT(*) as count
FROM tbl_upload_logs u
WHERE (? = '' OR u.module = ?)
	AND (? = '' OR u.summary LIKE '%' || ? || '%')
	AND (? = '' OR u.created_at >= ?)
	AND (? = '' OR u.created_at <= ?);

-- name: GetUploadLogsFiltered :many
SELECT
	u.id,
	u.module,
	u.outcome,
	u.kind,
	u.summary,
	u.filename,
	u.file_size,
	u.compress_preset,
	u.created_by,
	u.created_at,
	COALESCE(trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END), u.created_by_name, '') as created_by_name
FROM tbl_upload_logs u
LEFT JOIN tbl_teachers t ON u.created_by = t.id
WHERE (? = '' OR u.module = ?)
	AND (? = '' OR u.summary LIKE '%' || ? || '%')
	AND (? = '' OR u.created_at >= ?)
	AND (? = '' OR u.created_at <= ?)
ORDER BY u.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountUploadLogsByCreatedByFiltered :one
SELECT COUNT(*) as count
FROM tbl_upload_logs u
WHERE u.created_by = ?
	AND (? = '' OR u.module = ?)
	AND (? = '' OR u.summary LIKE '%' || ? || '%')
	AND (? = '' OR u.created_at >= ?)
	AND (? = '' OR u.created_at <= ?);

-- name: GetUploadLogsByCreatedByFiltered :many
SELECT
	u.id,
	u.module,
	u.outcome,
	u.kind,
	u.summary,
	u.filename,
	u.file_size,
	u.compress_preset,
	u.created_by,
	u.created_at,
	COALESCE(trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END), u.created_by_name, '') as created_by_name
FROM tbl_upload_logs u
LEFT JOIN tbl_teachers t ON u.created_by = t.id
WHERE u.created_by = ?
	AND (? = '' OR u.module = ?)
	AND (? = '' OR u.summary LIKE '%' || ? || '%')
	AND (? = '' OR u.created_at >= ?)
	AND (? = '' OR u.created_at <= ?)
ORDER BY u.created_at DESC
LIMIT ? OFFSET ?;
