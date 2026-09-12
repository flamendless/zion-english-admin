-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_training_materials (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	url TEXT NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('published', 'draft', 'deleted')) DEFAULT 'draft',
	created_by_id INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	deleted_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_training_materials_status ON tbl_training_materials(status);
CREATE INDEX IF NOT EXISTS idx_training_materials_created_by ON tbl_training_materials(created_by_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_training_materials_created_by;
DROP INDEX IF EXISTS idx_training_materials_status;
DROP TABLE IF EXISTS tbl_training_materials;
-- +goose StatementEnd
