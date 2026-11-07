-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_students ADD COLUMN legal_name TEXT CHECK (legal_name IS NULL OR length(legal_name) <= 64);
ALTER TABLE tbl_students ADD COLUMN nationality_id TEXT REFERENCES tbl_nationalities(id);
CREATE INDEX IF NOT EXISTS idx_students_nationality_id ON tbl_students(nationality_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_students_nationality_id;
ALTER TABLE tbl_students DROP COLUMN nationality_id;
ALTER TABLE tbl_students DROP COLUMN legal_name;
-- +goose StatementEnd
