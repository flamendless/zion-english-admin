-- +goose Up
ALTER TABLE tbl_teachers ADD COLUMN status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'approved'));

-- +goose Down
ALTER TABLE tbl_teachers DROP COLUMN status;
