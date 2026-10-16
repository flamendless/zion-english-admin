-- +goose Up
CREATE TABLE tbl_sql_console_saved_queries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	query_text TEXT NOT NULL,
	created_by INTEGER REFERENCES tbl_teachers(id),
	created_by_name TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE tbl_sql_console_query_log (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	query_text TEXT NOT NULL,
	saved_query_id INTEGER REFERENCES tbl_sql_console_saved_queries(id) ON DELETE SET NULL,
	outcome TEXT NOT NULL CHECK (outcome IN ('succeeded', 'failed', 'rejected')),
	error_message TEXT,
	duration_ms INTEGER,
	column_count INTEGER,
	rows_returned INTEGER,
	query_length INTEGER,
	truncated INTEGER NOT NULL DEFAULT 0,
	executed_by INTEGER REFERENCES tbl_teachers(id),
	executed_by_name TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_sql_console_query_log_created_at ON tbl_sql_console_query_log(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS tbl_sql_console_query_log;
DROP TABLE IF EXISTS tbl_sql_console_saved_queries;
