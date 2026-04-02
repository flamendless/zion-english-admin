-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    module TEXT NOT NULL,
    message TEXT NOT NULL,
    created_by INTEGER REFERENCES tbl_teachers(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_logs_id ON tbl_logs(id);
CREATE INDEX IF NOT EXISTS idx_logs_created_at ON tbl_logs(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS tbl_logs;
