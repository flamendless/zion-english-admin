-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_upload_logs_affiliate_kind (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	module TEXT NOT NULL,
	outcome TEXT NOT NULL CHECK (outcome IN ('succeeded', 'failed')),
	kind TEXT NOT NULL CHECK (kind IN ('intro_video', 'avatar', 'document', 'affiliate_csv', 'other')),
	summary TEXT NOT NULL,
	filename TEXT,
	file_size INTEGER,
	compress_preset TEXT CHECK (
		compress_preset IS NULL
		OR compress_preset IN ('very_low', 'medium', 'fine', 'high', 'original')
	),
	created_by INTEGER REFERENCES tbl_teachers(id),
	created_by_name TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO tbl_upload_logs_affiliate_kind (
	id,
	module,
	outcome,
	kind,
	summary,
	filename,
	file_size,
	compress_preset,
	created_by,
	created_by_name,
	created_at
)
SELECT
	id,
	module,
	outcome,
	kind,
	summary,
	filename,
	file_size,
	compress_preset,
	created_by,
	created_by_name,
	created_at
FROM tbl_upload_logs;

DROP TABLE tbl_upload_logs;
ALTER TABLE tbl_upload_logs_affiliate_kind RENAME TO tbl_upload_logs;

CREATE INDEX idx_upload_logs_created_at ON tbl_upload_logs(created_at DESC);
CREATE INDEX idx_upload_logs_created_by_created_at ON tbl_upload_logs(created_by, created_at DESC);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_upload_logs_legacy_kind (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	module TEXT NOT NULL,
	outcome TEXT NOT NULL CHECK (outcome IN ('succeeded', 'failed')),
	kind TEXT NOT NULL CHECK (kind IN ('intro_video', 'avatar', 'document', 'other')),
	summary TEXT NOT NULL,
	filename TEXT,
	file_size INTEGER,
	compress_preset TEXT CHECK (
		compress_preset IS NULL
		OR compress_preset IN ('very_low', 'medium', 'fine', 'high', 'original')
	),
	created_by INTEGER REFERENCES tbl_teachers(id),
	created_by_name TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO tbl_upload_logs_legacy_kind (
	id,
	module,
	outcome,
	kind,
	summary,
	filename,
	file_size,
	compress_preset,
	created_by,
	created_by_name,
	created_at
)
SELECT
	id,
	module,
	outcome,
	CASE WHEN kind = 'affiliate_csv' THEN 'other' ELSE kind END,
	summary,
	filename,
	file_size,
	compress_preset,
	created_by,
	created_by_name,
	created_at
FROM tbl_upload_logs;

DROP TABLE tbl_upload_logs;
ALTER TABLE tbl_upload_logs_legacy_kind RENAME TO tbl_upload_logs;

CREATE INDEX idx_upload_logs_created_at ON tbl_upload_logs(created_at DESC);
CREATE INDEX idx_upload_logs_created_by_created_at ON tbl_upload_logs(created_by, created_at DESC);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
