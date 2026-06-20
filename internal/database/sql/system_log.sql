-- name: GetAllLogs :many
SELECT l.id, l.module, l.message, l.created_by, l.created_at,
    COALESCE(t.name, l.created_by_name) as created_by_name
FROM tbl_logs l
LEFT JOIN tbl_teachers t ON l.created_by = t.id
ORDER BY l.created_at DESC;

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
