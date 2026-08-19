-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_report_generations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    output_path TEXT NOT NULL,
    record_count INTEGER NOT NULL,
    generated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE (teacher_id, start_date, end_date)
);

-- +goose Down
DROP TABLE IF EXISTS tbl_report_generations;
