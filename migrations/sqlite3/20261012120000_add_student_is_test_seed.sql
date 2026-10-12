-- +goose Up
ALTER TABLE tbl_students ADD COLUMN is_test_seed INTEGER NOT NULL DEFAULT 0 CHECK (is_test_seed IN (0, 1));

UPDATE tbl_students
SET is_test_seed = 1
WHERE name LIKE '%Seed Student%'
	OR name LIKE '%Test Student%';

CREATE INDEX IF NOT EXISTS idx_students_is_test_seed ON tbl_students(is_test_seed);

-- +goose Down
DROP INDEX IF EXISTS idx_students_is_test_seed;

-- SQLite cannot DROP COLUMN in older versions; recreate table without is_test_seed if needed.
-- goose Down: use table rebuild pattern from prior migrations.
CREATE TABLE tbl_students_drop_is_test_seed (
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
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO tbl_students_drop_is_test_seed (
	id, name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency,
	assigned_color, status, inactive_reason, deleted_reason, created_at, updated_at
)
SELECT
	id, name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency,
	assigned_color, status, inactive_reason, deleted_reason, created_at, updated_at
FROM tbl_students;

DROP TABLE tbl_students;
ALTER TABLE tbl_students_drop_is_test_seed RENAME TO tbl_students;

CREATE INDEX IF NOT EXISTS idx_students_status ON tbl_students(status);
CREATE INDEX IF NOT EXISTS idx_students_name ON tbl_students(name);
