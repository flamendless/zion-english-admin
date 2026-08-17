-- +goose Up
ALTER TABLE tbl_teachers ADD COLUMN deleted INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tbl_teachers ADD COLUMN deleted_at DATETIME;

DROP INDEX IF EXISTS idx_teachers_email;
DROP INDEX IF EXISTS idx_teachers_mobile_number;
CREATE UNIQUE INDEX idx_teachers_email ON tbl_teachers(email) WHERE deleted = 0;
CREATE UNIQUE INDEX idx_teachers_mobile_number ON tbl_teachers(mobile_number) WHERE deleted = 0;

-- +goose Down
DROP INDEX IF EXISTS idx_teachers_email;
DROP INDEX IF EXISTS idx_teachers_mobile_number;
CREATE UNIQUE INDEX idx_teachers_email ON tbl_teachers(email);
CREATE UNIQUE INDEX idx_teachers_mobile_number ON tbl_teachers(mobile_number);

ALTER TABLE tbl_teachers DROP COLUMN deleted_at;
ALTER TABLE tbl_teachers DROP COLUMN deleted;
