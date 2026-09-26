-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_meta_tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	content TEXT NOT NULL DEFAULT '',
	value TEXT NOT NULL DEFAULT '',
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_meta_tags_name ON tbl_meta_tags(name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_meta_tags_name;
DROP TABLE IF EXISTS tbl_meta_tags;
-- +goose StatementEnd
