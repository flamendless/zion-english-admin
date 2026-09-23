-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teacher_intro_videos_upload_source (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	original_filename TEXT,
	stored_filename TEXT,
	mime_type TEXT,
	file_size INTEGER,
	url TEXT,
	source_type TEXT CHECK(source_type IS NULL OR source_type IN ('upload', 'google_drive', 'youtube')),
	status TEXT NOT NULL CHECK(status IN ('submitted', 'approved', 'rejected', 'deleted')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	reviewed_at DATETIME,
	reviewed_by INTEGER,
	deleted_at DATETIME,
	reject_reason TEXT
);

INSERT INTO tbl_teacher_intro_videos_upload_source (
	id,
	teacher_id,
	original_filename,
	stored_filename,
	mime_type,
	file_size,
	url,
	source_type,
	status,
	created_at,
	reviewed_at,
	reviewed_by,
	deleted_at,
	reject_reason
)
SELECT
	id,
	teacher_id,
	original_filename,
	stored_filename,
	mime_type,
	file_size,
	url,
	source_type,
	status,
	created_at,
	reviewed_at,
	reviewed_by,
	deleted_at,
	reject_reason
FROM tbl_teacher_intro_videos;

DROP TABLE tbl_teacher_intro_videos;
ALTER TABLE tbl_teacher_intro_videos_upload_source RENAME TO tbl_teacher_intro_videos;

CREATE INDEX IF NOT EXISTS idx_teacher_intro_videos_teacher_id ON tbl_teacher_intro_videos(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_intro_videos_status ON tbl_teacher_intro_videos(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teacher_intro_videos_blocking
ON tbl_teacher_intro_videos(teacher_id)
WHERE status IN ('submitted', 'approved');

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teacher_intro_videos_link_source (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	original_filename TEXT,
	stored_filename TEXT,
	mime_type TEXT,
	file_size INTEGER,
	url TEXT,
	source_type TEXT CHECK(source_type IS NULL OR source_type IN ('google_drive', 'youtube')),
	status TEXT NOT NULL CHECK(status IN ('submitted', 'approved', 'rejected', 'deleted')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	reviewed_at DATETIME,
	reviewed_by INTEGER,
	deleted_at DATETIME,
	reject_reason TEXT
);

INSERT INTO tbl_teacher_intro_videos_link_source (
	id,
	teacher_id,
	original_filename,
	stored_filename,
	mime_type,
	file_size,
	url,
	source_type,
	status,
	created_at,
	reviewed_at,
	reviewed_by,
	deleted_at,
	reject_reason
)
SELECT
	id,
	teacher_id,
	original_filename,
	stored_filename,
	mime_type,
	file_size,
	url,
	source_type,
	status,
	created_at,
	reviewed_at,
	reviewed_by,
	deleted_at,
	reject_reason
FROM tbl_teacher_intro_videos
WHERE source_type IS NULL OR source_type IN ('google_drive', 'youtube');

DROP TABLE tbl_teacher_intro_videos;
ALTER TABLE tbl_teacher_intro_videos_link_source RENAME TO tbl_teacher_intro_videos;

CREATE INDEX IF NOT EXISTS idx_teacher_intro_videos_teacher_id ON tbl_teacher_intro_videos(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_intro_videos_status ON tbl_teacher_intro_videos(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teacher_intro_videos_blocking
ON tbl_teacher_intro_videos(teacher_id)
WHERE status IN ('submitted', 'approved');

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
