-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_useragents (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_agent TEXT NOT NULL UNIQUE,
	browser TEXT NOT NULL DEFAULT '',
	browser_version TEXT NOT NULL DEFAULT '',
	os TEXT NOT NULL DEFAULT '',
	device TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tbl_accesses (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL,
	login_at TEXT NOT NULL,
	logout_at TEXT,
	useragent_id INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	FOREIGN KEY (teacher_id) REFERENCES tbl_teachers(id),
	FOREIGN KEY (useragent_id) REFERENCES tbl_useragents(id)
);

CREATE INDEX IF NOT EXISTS idx_accesses_id ON tbl_accesses(id);

-- +goose Down
DROP INDEX IF EXISTS idx_accesses_id;
DROP TABLE IF EXISTS tbl_useragents;
DROP TABLE IF EXISTS tbl_accesses;
