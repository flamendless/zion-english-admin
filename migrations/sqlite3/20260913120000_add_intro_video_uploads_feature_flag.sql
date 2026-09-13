-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_feature_flags (key, enabled, visible_roles, value_text, updated_at)
VALUES ('intro_video.uploads', 0, '', '', datetime('now'))
ON CONFLICT(key) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_feature_flags WHERE key = 'intro_video.uploads';
-- +goose StatementEnd
