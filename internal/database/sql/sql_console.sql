-- name: ListSQLConsoleSavedQueries :many
SELECT
	id,
	name,
	query_text,
	created_by,
	created_by_name,
	created_at,
	updated_at
FROM tbl_sql_console_saved_queries
ORDER BY name ASC;

-- name: GetSQLConsoleSavedQueryByID :one
SELECT
	id,
	name,
	query_text,
	created_by,
	created_by_name,
	created_at,
	updated_at
FROM tbl_sql_console_saved_queries
WHERE id = ?;

-- name: InsertSQLConsoleSavedQuery :one
INSERT INTO tbl_sql_console_saved_queries (
	name,
	query_text,
	created_by,
	created_by_name,
	created_at,
	updated_at
) VALUES (
	?, ?, ?, ?, datetime('now'), datetime('now')
)
RETURNING id;

-- name: UpdateSQLConsoleSavedQuery :exec
UPDATE tbl_sql_console_saved_queries
SET
	name = ?,
	query_text = ?,
	updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteSQLConsoleSavedQuery :exec
DELETE FROM tbl_sql_console_saved_queries
WHERE id = ?;

-- name: InsertSQLConsoleQueryLog :one
INSERT INTO tbl_sql_console_query_log (
	query_text,
	saved_query_id,
	outcome,
	error_message,
	duration_ms,
	column_count,
	rows_returned,
	query_length,
	truncated,
	executed_by,
	executed_by_name,
	created_at
) VALUES (
	?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now')
)
RETURNING id;

-- name: CountSQLConsoleQueryLog :one
SELECT COUNT(*) AS count
FROM tbl_sql_console_query_log;

-- name: ListSQLConsoleQueryLog :many
SELECT
	id,
	query_text,
	saved_query_id,
	outcome,
	error_message,
	duration_ms,
	column_count,
	rows_returned,
	query_length,
	truncated,
	executed_by,
	executed_by_name,
	created_at
FROM tbl_sql_console_query_log
ORDER BY created_at DESC
LIMIT ? OFFSET ?;
