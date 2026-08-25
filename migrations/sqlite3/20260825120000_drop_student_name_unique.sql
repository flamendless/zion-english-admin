-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_students_no_name_unique (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    currency TEXT NOT NULL CHECK(currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
    contact TEXT,
    rate_per_class REAL NOT NULL,
    parent_name TEXT,
    assigned_color TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'inactive')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    inactive_reason TEXT
);

INSERT INTO tbl_students_no_name_unique (
    id, name, currency, contact, rate_per_class, parent_name, assigned_color,
    status, created_at, updated_at, inactive_reason
)
SELECT
    id, name, currency, contact, rate_per_class, parent_name, assigned_color,
    status, created_at, updated_at, inactive_reason
FROM tbl_students;

DROP TABLE tbl_students;
ALTER TABLE tbl_students_no_name_unique RENAME TO tbl_students;

CREATE INDEX IF NOT EXISTS idx_students_status ON tbl_students(status);
CREATE INDEX IF NOT EXISTS idx_students_name ON tbl_students(name);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_students_name_unique (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    currency TEXT NOT NULL CHECK(currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
    contact TEXT,
    rate_per_class REAL NOT NULL,
    parent_name TEXT,
    assigned_color TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'inactive')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    inactive_reason TEXT
);

INSERT INTO tbl_students_name_unique (
    id, name, currency, contact, rate_per_class, parent_name, assigned_color,
    status, created_at, updated_at, inactive_reason
)
SELECT
    id, name, currency, contact, rate_per_class, parent_name, assigned_color,
    status, created_at, updated_at, inactive_reason
FROM tbl_students;

DROP TABLE tbl_students;
ALTER TABLE tbl_students_name_unique RENAME TO tbl_students;

CREATE INDEX IF NOT EXISTS idx_students_status ON tbl_students(status);
CREATE INDEX IF NOT EXISTS idx_students_name ON tbl_students(name);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
