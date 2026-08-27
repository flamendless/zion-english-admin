-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_announcements (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	level TEXT NOT NULL CHECK(level IN ('info', 'warning', 'critical')),
	start_date TEXT NOT NULL,
	end_date TEXT NOT NULL,
	visible_to_all INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tbl_announcements_teachers_m2m (
	announcement_id INTEGER NOT NULL,
	teacher_id INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (announcement_id, teacher_id),
	FOREIGN KEY (announcement_id) REFERENCES tbl_announcements(id) ON DELETE CASCADE,
	FOREIGN KEY (teacher_id) REFERENCES tbl_teachers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_announcements_dates ON tbl_announcements(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_announcements_teachers_m2m_teacher ON tbl_announcements_teachers_m2m(teacher_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_announcements_teachers_m2m_teacher;
DROP INDEX IF EXISTS idx_announcements_dates;
DROP TABLE IF EXISTS tbl_announcements_teachers_m2m;
DROP TABLE IF EXISTS tbl_announcements;
-- +goose StatementEnd
