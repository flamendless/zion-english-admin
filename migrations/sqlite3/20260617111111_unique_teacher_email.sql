-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS idx_teachers_email ON tbl_teachers(email);

-- +goose Down
DROP INDEX IF EXISTS idx_teachers_email;
