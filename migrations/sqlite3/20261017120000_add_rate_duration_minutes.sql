-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_teachers ADD COLUMN rate_duration_minutes INTEGER NOT NULL DEFAULT 25;
UPDATE tbl_teachers SET rate_duration_minutes = 25 WHERE rate_duration_minutes IS NULL;

ALTER TABLE tbl_students ADD COLUMN rate_duration_minutes INTEGER NOT NULL DEFAULT 25;
UPDATE tbl_students SET rate_duration_minutes = 25 WHERE rate_duration_minutes IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_students DROP COLUMN rate_duration_minutes;
ALTER TABLE tbl_teachers DROP COLUMN rate_duration_minutes;
-- +goose StatementEnd
