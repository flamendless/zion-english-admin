-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_teachers_students_m2m (
	teacher_id INTEGER NOT NULL,
	student_id INTEGER NOT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (teacher_id, student_id),
	FOREIGN KEY (teacher_id) REFERENCES tbl_teachers(id) ON DELETE CASCADE,
	FOREIGN KEY (student_id) REFERENCES tbl_students(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_teachers_students_m2m_teacher_id ON tbl_teachers_students_m2m(teacher_id);
CREATE INDEX IF NOT EXISTS idx_teachers_students_m2m_student_id ON tbl_teachers_students_m2m(student_id);

-- +goose Down
DROP INDEX IF EXISTS idx_teachers_students_m2m_teacher_id;
DROP INDEX IF EXISTS idx_teachers_students_m2m_student_id;
DROP TABLE IF EXISTS tbl_teachers_students_m2m;
