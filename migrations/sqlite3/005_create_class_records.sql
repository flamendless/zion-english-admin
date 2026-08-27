-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_class_records (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	student_id INTEGER NOT NULL REFERENCES tbl_students(id),
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	date TEXT NOT NULL,
	duration_minutes INTEGER NOT NULL,
	rate REAL NOT NULL,
	currency TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'conducted',
	reason TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	recorded_by_role TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS tbl_class_records;
