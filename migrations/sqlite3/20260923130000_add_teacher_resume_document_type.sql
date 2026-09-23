-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teacher_documents_resume_type (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	type TEXT NOT NULL CHECK(type IN ('avatar', 'document', 'resume')),
	original_filename TEXT NOT NULL,
	stored_filename TEXT NOT NULL,
	file_extension TEXT NOT NULL,
	file_size INTEGER NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('submitted', 'approved', 'rejected')),
	uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	reviewed_at DATETIME,
	reviewed_by INTEGER
);

INSERT INTO tbl_teacher_documents_resume_type (
	id, teacher_id, type, original_filename, stored_filename, file_extension, file_size, status, uploaded_at, reviewed_at, reviewed_by
)
SELECT
	id, teacher_id, type, original_filename, stored_filename, file_extension, file_size, status, uploaded_at, reviewed_at, reviewed_by
FROM tbl_teacher_documents;

DROP TABLE tbl_teacher_documents;
ALTER TABLE tbl_teacher_documents_resume_type RENAME TO tbl_teacher_documents;

CREATE INDEX IF NOT EXISTS idx_teacher_documents_teacher_id ON tbl_teacher_documents(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_documents_status ON tbl_teacher_documents(status);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teacher_documents_legacy (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	type TEXT NOT NULL CHECK(type IN ('avatar', 'document')),
	original_filename TEXT NOT NULL,
	stored_filename TEXT NOT NULL,
	file_extension TEXT NOT NULL,
	file_size INTEGER NOT NULL,
	status TEXT NOT NULL CHECK(status IN ('submitted', 'approved', 'rejected')),
	uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	reviewed_at DATETIME,
	reviewed_by INTEGER
);

INSERT INTO tbl_teacher_documents_legacy (
	id, teacher_id, type, original_filename, stored_filename, file_extension, file_size, status, uploaded_at, reviewed_at, reviewed_by
)
SELECT
	id, teacher_id, type, original_filename, stored_filename, file_extension, file_size, status, uploaded_at, reviewed_at, reviewed_by
FROM tbl_teacher_documents
WHERE type IN ('avatar', 'document');

DROP TABLE tbl_teacher_documents;
ALTER TABLE tbl_teacher_documents_legacy RENAME TO tbl_teacher_documents;

CREATE INDEX IF NOT EXISTS idx_teacher_documents_teacher_id ON tbl_teacher_documents(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_documents_status ON tbl_teacher_documents(status);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
