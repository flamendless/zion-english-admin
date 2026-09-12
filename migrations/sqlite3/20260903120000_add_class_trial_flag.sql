-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_class_records ADD COLUMN is_trial_class INTEGER NOT NULL DEFAULT 0;
ALTER TABLE tbl_scheduled_classes ADD COLUMN is_trial_class INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_scheduled_classes DROP COLUMN is_trial_class;
ALTER TABLE tbl_class_records DROP COLUMN is_trial_class;
-- +goose StatementEnd
