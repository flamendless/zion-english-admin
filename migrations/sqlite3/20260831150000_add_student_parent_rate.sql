-- +goose Up
ALTER TABLE tbl_students ADD COLUMN parent_rate REAL;
ALTER TABLE tbl_students ADD COLUMN parent_currency TEXT CHECK(parent_currency IS NULL OR parent_currency IN ('KRW', 'CAD', 'YEN', 'PHP'));

-- +goose Down
ALTER TABLE tbl_students DROP COLUMN parent_currency;
ALTER TABLE tbl_students DROP COLUMN parent_rate;
