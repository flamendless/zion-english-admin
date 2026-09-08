-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_scheduled_class_series (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	created_by_role TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

ALTER TABLE tbl_scheduled_classes ADD COLUMN series_id INTEGER REFERENCES tbl_scheduled_class_series(id);

CREATE INDEX IF NOT EXISTS idx_scheduled_classes_series_date
	ON tbl_scheduled_classes(series_id, scheduled_date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_scheduled_classes_series_date;
ALTER TABLE tbl_scheduled_classes DROP COLUMN series_id;
DROP TABLE IF EXISTS tbl_scheduled_class_series;
-- +goose StatementEnd
