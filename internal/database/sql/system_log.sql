-- name: GetAllLogs :many
SELECT l.id, l.module, l.message, l.created_by, l.created_at,
    COALESCE(trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END), l.created_by_name, '') as created_by_name
FROM tbl_logs l
LEFT JOIN tbl_teachers t ON l.created_by = t.id
ORDER BY l.created_at DESC;

-- name: GetLogsByCreatedBy :many
SELECT l.id, l.module, l.message, l.created_by, l.created_at,
    COALESCE(trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END), l.created_by_name, '') as created_by_name
FROM tbl_logs l
LEFT JOIN tbl_teachers t ON l.created_by = t.id
WHERE l.created_by = ?
ORDER BY l.created_at DESC;

-- name: CountAllLogsFiltered :one
SELECT COUNT(*) as count
FROM tbl_logs l
WHERE (? = '' OR l.module = ?)
  AND (? = '' OR l.message LIKE '%' || ? || '%')
  AND (? = '' OR l.created_at >= ?)
  AND (? = '' OR l.created_at <= ?);

-- name: GetAllLogsFiltered :many
SELECT l.id, l.module, l.message, l.created_by, l.created_at,
    COALESCE(trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END), l.created_by_name, '') as created_by_name
FROM tbl_logs l
LEFT JOIN tbl_teachers t ON l.created_by = t.id
WHERE (? = '' OR l.module = ?)
  AND (? = '' OR l.message LIKE '%' || ? || '%')
  AND (? = '' OR l.created_at >= ?)
  AND (? = '' OR l.created_at <= ?)
ORDER BY l.created_at DESC
LIMIT ? OFFSET ?;

-- name: CountLogsByCreatedByFiltered :one
SELECT COUNT(*) as count
FROM tbl_logs l
WHERE l.created_by = ?
  AND (? = '' OR l.module = ?)
  AND (? = '' OR l.message LIKE '%' || ? || '%')
  AND (? = '' OR l.created_at >= ?)
  AND (? = '' OR l.created_at <= ?);

-- name: GetLogsByCreatedByFiltered :many
SELECT l.id, l.module, l.message, l.created_by, l.created_at,
    COALESCE(trim(t.first_name || CASE WHEN t.middle_name != '' THEN ' ' || t.middle_name ELSE '' END || CASE WHEN t.last_name != '' THEN ' ' || t.last_name ELSE '' END), l.created_by_name, '') as created_by_name
FROM tbl_logs l
LEFT JOIN tbl_teachers t ON l.created_by = t.id
WHERE l.created_by = ?
  AND (? = '' OR l.module = ?)
  AND (? = '' OR l.message LIKE '%' || ? || '%')
  AND (? = '' OR l.created_at >= ?)
  AND (? = '' OR l.created_at <= ?)
ORDER BY l.created_at DESC
LIMIT ? OFFSET ?;

-- name: InsertLog :exec
INSERT INTO tbl_logs (
    module,
    message,
    created_by,
    created_by_name,
    created_at
) VALUES (
    ?, ?, ?, ?, datetime('now')
);
