-- +goose Up
ALTER TABLE tbl_students ADD COLUMN student_number TEXT;
ALTER TABLE tbl_students ADD COLUMN parent_number TEXT;

UPDATE tbl_students
SET student_number = substr(
	substr(name, instr(name, '(Z') + 1),
	1,
	instr(substr(name, instr(name, '(Z') + 1), ')') - 1
)
WHERE instr(name, '(Z') > 0
	AND instr(substr(name, instr(name, '(Z') + 1), ')') > 1;

UPDATE tbl_students
SET name = trim(substr(name, 1, instr(name, '(Z') - 1))
WHERE student_number IS NOT NULL
	AND trim(student_number) != ''
	AND instr(name, '(Z') > 0;

-- +goose Down
CREATE TABLE tbl_students_drop_numbers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	currency TEXT NOT NULL CHECK(currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
	contact TEXT,
	rate_per_class REAL NOT NULL,
	parent_name TEXT,
	parent_rate REAL,
	parent_currency TEXT CHECK(parent_currency IS NULL OR parent_currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
	assigned_color TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'inactive', 'deleted')),
	inactive_reason TEXT,
	deleted_reason TEXT,
	is_test_seed INTEGER NOT NULL DEFAULT 0 CHECK (is_test_seed IN (0, 1)),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO tbl_students_drop_numbers (
	id, name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency,
	assigned_color, status, inactive_reason, deleted_reason, is_test_seed, created_at, updated_at
)
SELECT
	id, name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency,
	assigned_color, status, inactive_reason, deleted_reason, is_test_seed, created_at, updated_at
FROM tbl_students;

DROP TABLE tbl_students;
ALTER TABLE tbl_students_drop_numbers RENAME TO tbl_students;

CREATE INDEX IF NOT EXISTS idx_students_status ON tbl_students(status);
CREATE INDEX IF NOT EXISTS idx_students_name ON tbl_students(name);
CREATE INDEX IF NOT EXISTS idx_students_is_test_seed ON tbl_students(is_test_seed);
