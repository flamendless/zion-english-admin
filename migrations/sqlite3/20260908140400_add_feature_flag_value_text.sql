-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_feature_flags ADD COLUMN value_text TEXT NOT NULL DEFAULT '';

INSERT INTO tbl_feature_flags (key, enabled, visible_roles, value_text, updated_at)
VALUES ('class.overdue_grace_period_minutes', 1, '', '0', datetime('now'))
ON CONFLICT(key) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_feature_flags WHERE key = 'class.overdue_grace_period_minutes';

ALTER TABLE tbl_feature_flags DROP COLUMN value_text;
-- +goose StatementEnd
