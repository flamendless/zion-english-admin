-- +goose Up
CREATE TABLE tbl_upload_logs (
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

CREATE INDEX idx_upload_logs_created_at ON tbl_upload_logs(created_at DESC);
CREATE INDEX idx_upload_logs_created_by_created_at ON tbl_upload_logs(created_by, created_at DESC);

INSERT INTO tbl_upload_logs (
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
	l.module,
	CASE
		WHEN l.message LIKE 'upload error:%' THEN 'failed'
		WHEN l.message LIKE 'upload log: succeeded:%' THEN 'succeeded'
		WHEN l.message LIKE 'upload log: failed:%' THEN 'failed'
		ELSE 'failed'
	END,
	CASE
		WHEN l.message LIKE '%intro video%' THEN 'intro_video'
		WHEN l.message LIKE '%profile picture%' THEN 'avatar'
		WHEN l.message LIKE '%ID document%' OR l.message LIKE '%resume/CV%' THEN 'document'
		ELSE 'other'
	END,
	CASE
		WHEN l.message LIKE 'upload log: succeeded: %' THEN substr(l.message, 24)
		WHEN l.message LIKE 'upload log: failed: %' THEN substr(l.message, 20)
		WHEN l.message LIKE 'upload error: %' THEN substr(l.message, 15)
		ELSE l.message
	END,
	CASE
		WHEN instr(l.message, 'file ''') > 0 THEN substr(
			l.message,
			instr(l.message, 'file ''') + 6,
			instr(substr(l.message, instr(l.message, 'file ''') + 6), '''') - 1
		)
		ELSE NULL
	END,
	NULL,
	NULL,
	l.created_by,
	l.created_by_name,
	l.created_at
FROM tbl_logs l
WHERE l.message LIKE 'upload log:%'
	OR l.message LIKE 'upload error:%';

DELETE FROM tbl_logs
WHERE message LIKE 'upload log:%'
	OR message LIKE 'upload error:%';

-- +goose Down
INSERT INTO tbl_logs (module, message, created_by, created_by_name, created_at)
SELECT
	u.module,
	'upload log: ' || u.outcome || ': ' || u.summary,
	u.created_by,
	u.created_by_name,
	u.created_at
FROM tbl_upload_logs u;

DROP TABLE IF EXISTS tbl_upload_logs;
