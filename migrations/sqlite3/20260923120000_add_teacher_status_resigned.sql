-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teachers_resigned_status (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
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
	status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'approved', 'resigned')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	profile_picture TEXT,
	password_changed_at DATETIME,
	mobile_changed_at DATETIME,
	deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME
);

INSERT INTO tbl_teachers_resigned_status (
	id, first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class, currency,
	drive_url, sex, password, template, status, created_at, updated_at,
	profile_picture, password_changed_at, mobile_changed_at, deleted, deleted_at
)
SELECT
	id, first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class, currency,
	drive_url, sex, password, template, status, created_at, updated_at,
	profile_picture, password_changed_at, mobile_changed_at, deleted, deleted_at
FROM tbl_teachers;

DROP TABLE tbl_teachers;
ALTER TABLE tbl_teachers_resigned_status RENAME TO tbl_teachers;

CREATE INDEX IF NOT EXISTS idx_teachers_status ON tbl_teachers(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_email ON tbl_teachers(email) WHERE deleted = 0;
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_mobile_number ON tbl_teachers(mobile_number) WHERE deleted = 0;

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teachers_legacy_status (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
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
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	profile_picture TEXT,
	password_changed_at DATETIME,
	mobile_changed_at DATETIME,
	deleted INTEGER NOT NULL DEFAULT 0,
	deleted_at DATETIME
);

INSERT INTO tbl_teachers_legacy_status (
	id, first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class, currency,
	drive_url, sex, password, template, status, created_at, updated_at,
	profile_picture, password_changed_at, mobile_changed_at, deleted, deleted_at
)
SELECT
	id, first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class, currency,
	drive_url, sex, password, template,
	CASE WHEN status = 'resigned' THEN 'approved' ELSE status END,
	created_at, updated_at,
	profile_picture, password_changed_at, mobile_changed_at, deleted, deleted_at
FROM tbl_teachers;

DROP TABLE tbl_teachers;
ALTER TABLE tbl_teachers_legacy_status RENAME TO tbl_teachers;

CREATE INDEX IF NOT EXISTS idx_teachers_status ON tbl_teachers(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_email ON tbl_teachers(email) WHERE deleted = 0;
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_mobile_number ON tbl_teachers(mobile_number) WHERE deleted = 0;

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
