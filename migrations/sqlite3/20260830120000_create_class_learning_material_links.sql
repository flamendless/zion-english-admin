-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_class_record_learning_materials (
	class_record_id INTEGER NOT NULL REFERENCES tbl_class_records(id),
	material_id INTEGER NOT NULL REFERENCES tbl_learning_materials(id),
	created_at DATETIME NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (class_record_id, material_id)
);

CREATE INDEX idx_class_record_lm_record ON tbl_class_record_learning_materials(class_record_id);

CREATE TABLE tbl_scheduled_class_learning_materials (
	scheduled_class_id INTEGER NOT NULL REFERENCES tbl_scheduled_classes(id),
	material_id INTEGER NOT NULL REFERENCES tbl_learning_materials(id),
	created_at DATETIME NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (scheduled_class_id, material_id)
);

CREATE INDEX idx_scheduled_class_lm_schedule ON tbl_scheduled_class_learning_materials(scheduled_class_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_scheduled_class_lm_schedule;
DROP TABLE IF EXISTS tbl_scheduled_class_learning_materials;
DROP INDEX IF EXISTS idx_class_record_lm_record;
DROP TABLE IF EXISTS tbl_class_record_learning_materials;
-- +goose StatementEnd
