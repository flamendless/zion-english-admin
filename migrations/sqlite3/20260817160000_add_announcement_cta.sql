-- +goose Up
ALTER TABLE tbl_announcements ADD COLUMN cta_label TEXT NOT NULL DEFAULT '';
ALTER TABLE tbl_announcements ADD COLUMN cta_url TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE tbl_announcements DROP COLUMN cta_url;
ALTER TABLE tbl_announcements DROP COLUMN cta_label;
