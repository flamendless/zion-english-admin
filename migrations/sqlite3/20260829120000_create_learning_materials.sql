-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_learning_material_tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	label TEXT NOT NULL UNIQUE COLLATE NOCASE,
	color TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tbl_learning_materials (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	owner_id INTEGER NOT NULL DEFAULT 0,
	description TEXT NOT NULL,
	url TEXT NOT NULL,
	access TEXT NOT NULL CHECK(access IN ('public', 'private')) DEFAULT 'public',
	status TEXT NOT NULL CHECK(status IN ('published', 'draft', 'deleted')) DEFAULT 'draft',
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	deleted_at TEXT
);

CREATE TABLE IF NOT EXISTS tbl_learning_materials_tags_m2m (
	material_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (material_id, tag_id),
	FOREIGN KEY (material_id) REFERENCES tbl_learning_materials(id) ON DELETE CASCADE,
	FOREIGN KEY (tag_id) REFERENCES tbl_learning_material_tags(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_learning_materials_owner ON tbl_learning_materials(owner_id);
CREATE INDEX IF NOT EXISTS idx_learning_materials_status ON tbl_learning_materials(status);
CREATE INDEX IF NOT EXISTS idx_learning_materials_access ON tbl_learning_materials(access);
CREATE INDEX IF NOT EXISTS idx_learning_material_tags_label ON tbl_learning_material_tags(label);
CREATE INDEX IF NOT EXISTS idx_learning_materials_tags_m2m_tag ON tbl_learning_materials_tags_m2m(tag_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_learning_materials_tags_m2m_tag;
DROP INDEX IF EXISTS idx_learning_material_tags_label;
DROP INDEX IF EXISTS idx_learning_materials_access;
DROP INDEX IF EXISTS idx_learning_materials_status;
DROP INDEX IF EXISTS idx_learning_materials_owner;
DROP TABLE IF EXISTS tbl_learning_materials_tags_m2m;
DROP TABLE IF EXISTS tbl_learning_materials;
DROP TABLE IF EXISTS tbl_learning_material_tags;
-- +goose StatementEnd
