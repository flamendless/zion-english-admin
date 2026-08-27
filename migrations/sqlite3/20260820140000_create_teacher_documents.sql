-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_teacher_documents (
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

CREATE INDEX IF NOT EXISTS idx_teacher_documents_teacher_id ON tbl_teacher_documents(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teacher_documents_status ON tbl_teacher_documents(status);

-- +goose Down
DROP INDEX IF EXISTS idx_teacher_documents_status;
DROP INDEX IF EXISTS idx_teacher_documents_teacher_id;
DROP TABLE IF EXISTS tbl_teacher_documents;
