-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teachers_dedup_name (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	first_name TEXT NOT NULL DEFAULT '',
	middle_name TEXT NOT NULL DEFAULT '',
	last_name TEXT NOT NULL DEFAULT '',
	birthdate TEXT NOT NULL,
	address TEXT NOT NULL,
	joining_date TEXT NOT NULL,
	mobile_number TEXT NOT NULL,
	email TEXT NOT NULL,
	certifications TEXT,
	assigned_color TEXT NOT NULL,
	rate_per_class REAL NOT NULL,
	currency TEXT NOT NULL CHECK(currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
	drive_url TEXT NOT NULL,
	sex TEXT CHECK(sex IN ('M', 'F')),
	password TEXT NOT NULL DEFAULT '',
	template TEXT,
	status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'approved')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO tbl_teachers_dedup_name (
	id, name, first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class, currency,
	drive_url, sex, password, template, status, created_at, updated_at
)
SELECT
	id, name, first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class, currency,
	drive_url, sex, password, template, status, created_at, updated_at
FROM tbl_teachers;

DROP TABLE tbl_teachers;
ALTER TABLE tbl_teachers_dedup_name RENAME TO tbl_teachers;

CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_email ON tbl_teachers(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_mobile_number ON tbl_teachers(mobile_number);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teachers_unique_name (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	first_name TEXT NOT NULL DEFAULT '',
	middle_name TEXT NOT NULL DEFAULT '',
	last_name TEXT NOT NULL DEFAULT '',
	birthdate TEXT NOT NULL,
	address TEXT NOT NULL,
	joining_date TEXT NOT NULL,
	mobile_number TEXT NOT NULL,
	email TEXT NOT NULL,
	certifications TEXT,
	assigned_color TEXT NOT NULL,
	rate_per_class REAL NOT NULL,
	currency TEXT NOT NULL CHECK(currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
	drive_url TEXT NOT NULL,
	sex TEXT CHECK(sex IN ('M', 'F')),
	password TEXT NOT NULL DEFAULT '',
	template TEXT,
	status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'approved')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO tbl_teachers_unique_name (
	id, name, first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class, currency,
	drive_url, sex, password, template, status, created_at, updated_at
)
SELECT
	id, name, first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class, currency,
	drive_url, sex, password, template, status, created_at, updated_at
FROM tbl_teachers;

DROP TABLE tbl_teachers;
ALTER TABLE tbl_teachers_unique_name RENAME TO tbl_teachers;

CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_email ON tbl_teachers(email);
DROP INDEX IF EXISTS idx_teachers_mobile_number;

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
