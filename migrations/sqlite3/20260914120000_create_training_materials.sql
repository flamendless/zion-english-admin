-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_training_material_tags (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	label TEXT NOT NULL UNIQUE COLLATE NOCASE,
	color TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tbl_training_materials (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	url TEXT NOT NULL,
	embed_url TEXT NOT NULL,
	source_type TEXT NOT NULL CHECK(source_type IN ('youtube')) DEFAULT 'youtube',
	video_id TEXT NOT NULL,
	thumbnail_url TEXT NOT NULL DEFAULT '',
	duration_seconds INTEGER,
	status TEXT NOT NULL CHECK(status IN ('published', 'draft', 'deleted')) DEFAULT 'draft',
	created_by INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	deleted_at TEXT,
	FOREIGN KEY (created_by) REFERENCES tbl_teachers(id)
);

CREATE TABLE IF NOT EXISTS tbl_training_materials_tags_m2m (
	material_id INTEGER NOT NULL,
	tag_id INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (material_id, tag_id),
	FOREIGN KEY (material_id) REFERENCES tbl_training_materials(id) ON DELETE CASCADE,
	FOREIGN KEY (tag_id) REFERENCES tbl_training_material_tags(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tbl_training_material_progress (
	material_id INTEGER NOT NULL,
	teacher_id INTEGER NOT NULL,
	watch_seconds INTEGER NOT NULL DEFAULT 0,
	progress_percent REAL NOT NULL DEFAULT 0,
	completed_at TEXT,
	first_viewed_at TEXT NOT NULL DEFAULT (datetime('now')),
	last_viewed_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (material_id, teacher_id),
	FOREIGN KEY (material_id) REFERENCES tbl_training_materials(id) ON DELETE CASCADE,
	FOREIGN KEY (teacher_id) REFERENCES tbl_teachers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_training_materials_status ON tbl_training_materials(status);
CREATE INDEX IF NOT EXISTS idx_training_materials_source_type ON tbl_training_materials(source_type);
CREATE INDEX IF NOT EXISTS idx_training_material_tags_label ON tbl_training_material_tags(label);
CREATE INDEX IF NOT EXISTS idx_training_materials_tags_m2m_tag ON tbl_training_materials_tags_m2m(tag_id);
CREATE INDEX IF NOT EXISTS idx_training_material_progress_teacher ON tbl_training_material_progress(teacher_id);
CREATE INDEX IF NOT EXISTS idx_training_material_progress_material ON tbl_training_material_progress(material_id);
CREATE INDEX IF NOT EXISTS idx_training_material_progress_completed ON tbl_training_material_progress(completed_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_training_material_progress_completed;
DROP INDEX IF EXISTS idx_training_material_progress_material;
DROP INDEX IF EXISTS idx_training_material_progress_teacher;
DROP INDEX IF EXISTS idx_training_materials_tags_m2m_tag;
DROP INDEX IF EXISTS idx_training_material_tags_label;
DROP INDEX IF EXISTS idx_training_materials_source_type;
DROP INDEX IF EXISTS idx_training_materials_status;
DROP TABLE IF EXISTS tbl_training_material_progress;
DROP TABLE IF EXISTS tbl_training_materials_tags_m2m;
DROP TABLE IF EXISTS tbl_training_materials;
DROP TABLE IF EXISTS tbl_training_material_tags;
-- +goose StatementEnd
