-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_feature_flags (
	key TEXT PRIMARY KEY,
	enabled INTEGER NOT NULL DEFAULT 1,
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_feature_flags;
-- +goose StatementEnd
