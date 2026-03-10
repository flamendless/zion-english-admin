-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_teachers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    birthdate TEXT NOT NULL,
    address TEXT NOT NULL,
    joining_date TEXT NOT NULL,
    mobile_number TEXT NOT NULL,
    email TEXT NOT NULL,
    certifications TEXT,
    assigned_color TEXT NOT NULL,
    rate_per_class REAL NOT NULL,
    currency TEXT NOT NULL CHECK(currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tbl_teachers;
-- +goose StatementEnd
