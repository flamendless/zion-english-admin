-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_feature_flags (key, enabled, visible_roles, value_text, updated_at)
VALUES ('ads.display', 1, '', '', datetime('now'))
ON CONFLICT(key) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_feature_flags WHERE key = 'ads.display';
-- +goose StatementEnd
