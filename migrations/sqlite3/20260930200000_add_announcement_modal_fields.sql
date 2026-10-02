-- +goose Up
ALTER TABLE tbl_announcements ADD COLUMN display_type TEXT NOT NULL DEFAULT 'banner';
ALTER TABLE tbl_announcements ADD COLUMN modal_frequency TEXT;
ALTER TABLE tbl_announcements ADD COLUMN repeat_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tbl_announcements ADD COLUMN repeat_schedule TEXT;
ALTER TABLE tbl_announcements ADD COLUMN cutoff_repeat_days INTEGER;

-- +goose Down
ALTER TABLE tbl_announcements DROP COLUMN cutoff_repeat_days;
ALTER TABLE tbl_announcements DROP COLUMN repeat_schedule;
ALTER TABLE tbl_announcements DROP COLUMN repeat_enabled;
ALTER TABLE tbl_announcements DROP COLUMN modal_frequency;
ALTER TABLE tbl_announcements DROP COLUMN display_type;
