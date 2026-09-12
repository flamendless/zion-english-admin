-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_teacher_intro_videos (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	original_filename TEXT NOT NULL,
	stored_filename TEXT NOT NULL,
	mime_type TEXT NOT NULL,
	file_size INTEGER NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('submitted', 'approved', 'rejected', 'deleted')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	reviewed_at DATETIME,
	reviewed_by INTEGER,
	deleted_at DATETIME,
	reject_reason TEXT
);

CREATE INDEX IF NOT EXISTS idx_teacher_intro_videos_teacher_id ON tbl_teacher_intro_videos(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_intro_videos_status ON tbl_teacher_intro_videos(status);

-- +goose Down
DROP INDEX IF EXISTS idx_teacher_intro_videos_status;
DROP INDEX IF EXISTS idx_teacher_intro_videos_teacher_id;
DROP TABLE IF EXISTS tbl_teacher_intro_videos;
