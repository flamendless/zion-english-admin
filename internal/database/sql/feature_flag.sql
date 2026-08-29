-- name: GetFeatureFlag :one
SELECT key, enabled, updated_at
FROM tbl_feature_flags
WHERE key = ?;

-- name: ListFeatureFlags :many
SELECT key, enabled, updated_at
FROM tbl_feature_flags
ORDER BY key ASC;

-- name: UpsertFeatureFlag :exec
INSERT INTO tbl_feature_flags (key, enabled, updated_at)
VALUES (?, ?, datetime('now'))
ON CONFLICT(key) DO UPDATE SET
	enabled = excluded.enabled,
	updated_at = datetime('now');
