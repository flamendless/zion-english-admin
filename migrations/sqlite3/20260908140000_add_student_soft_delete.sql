-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_students_soft_delete (
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
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	inactive_reason TEXT,
	deleted_reason TEXT
);

INSERT INTO tbl_students_soft_delete (
	id, name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency,
	assigned_color, status, created_at, updated_at, inactive_reason, deleted_reason
)
SELECT
	id, name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency,
	assigned_color, status, created_at, updated_at, inactive_reason, NULL
FROM tbl_students;

DROP TABLE tbl_students;
ALTER TABLE tbl_students_soft_delete RENAME TO tbl_students;

CREATE INDEX IF NOT EXISTS idx_students_status ON tbl_students(status);
CREATE INDEX IF NOT EXISTS idx_students_name ON tbl_students(name);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_students_no_soft_delete (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	currency TEXT NOT NULL CHECK(currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
	contact TEXT,
	rate_per_class REAL NOT NULL,
	parent_name TEXT,
	parent_rate REAL,
	parent_currency TEXT CHECK(parent_currency IS NULL OR parent_currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
	assigned_color TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'inactive')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	inactive_reason TEXT
);

INSERT INTO tbl_students_no_soft_delete (
	id, name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency,
	assigned_color, status, created_at, updated_at, inactive_reason
)
SELECT
	id, name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency,
	assigned_color, status, created_at, updated_at, inactive_reason
FROM tbl_students
WHERE status != 'deleted';

DROP TABLE tbl_students;
ALTER TABLE tbl_students_no_soft_delete RENAME TO tbl_students;

CREATE INDEX IF NOT EXISTS idx_students_status ON tbl_students(status);
CREATE INDEX IF NOT EXISTS idx_students_name ON tbl_students(name);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
