-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_student_relationships (
	student_id INTEGER NOT NULL,
	related_student_id INTEGER NOT NULL,
	relationship TEXT,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (student_id, related_student_id),
	FOREIGN KEY (student_id) REFERENCES tbl_students(id) ON DELETE CASCADE,
	FOREIGN KEY (related_student_id) REFERENCES tbl_students(id) ON DELETE CASCADE,
	CHECK (student_id != related_student_id)
);

CREATE INDEX IF NOT EXISTS idx_student_relationships_student_id ON tbl_student_relationships(student_id);
CREATE INDEX IF NOT EXISTS idx_student_relationships_related_student_id ON tbl_student_relationships(related_student_id);

-- +goose Down
DROP INDEX IF EXISTS idx_student_relationships_related_student_id;
DROP INDEX IF EXISTS idx_student_relationships_student_id;
DROP TABLE IF EXISTS tbl_student_relationships;
