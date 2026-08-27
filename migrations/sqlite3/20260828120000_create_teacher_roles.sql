-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_teacher_roles (
	teacher_id INTEGER NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('teacher', 'admin', 'developer')),
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (teacher_id, role),
	FOREIGN KEY (teacher_id) REFERENCES tbl_teachers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_teacher_roles_teacher_id ON tbl_teacher_roles(teacher_id);

INSERT INTO tbl_teacher_roles (teacher_id, role)
SELECT id, 'teacher' FROM tbl_teachers;

-- +goose Down
DROP INDEX IF EXISTS idx_teacher_roles_teacher_id;
DROP TABLE IF EXISTS tbl_teacher_roles;
