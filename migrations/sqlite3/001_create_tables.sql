-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_processing_logs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	google_drive_url TEXT NOT NULL,
	name TEXT NOT NULL,
	template TEXT,
	start_date TEXT NOT NULL,
	end_date TEXT NOT NULL,
	excluded_rows TEXT,
	useragent TEXT,
	output_path TEXT,
	errors TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tbl_records (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	google_drive_url TEXT NOT NULL,
	student_name TEXT NOT NULL,
	date TEXT NOT NULL,
	duration_minutes NUMBER NOT NULL,
	rate NUMBER NOT NULL,
	status TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_processing_logs;
DROP TABLE IF EXISTS tbl_records;
-- +goose StatementEnd
