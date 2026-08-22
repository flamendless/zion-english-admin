-- +goose Up
ALTER TABLE tbl_announcements
  ADD COLUMN status TEXT NOT NULL DEFAULT 'published'
  CHECK(status IN ('published', 'draft', 'deleted'));

-- +goose Down
ALTER TABLE tbl_announcements DROP COLUMN status;
