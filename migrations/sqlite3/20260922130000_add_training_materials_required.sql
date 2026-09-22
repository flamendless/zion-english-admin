-- +goose Up
-- +goose StatementBegin
-- Skip when required already exists (e.g. manual dev apply before goose ran this version).
ALTER TABLE tbl_training_materials ADD COLUMN required INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_training_materials DROP COLUMN required;
-- +goose StatementEnd
