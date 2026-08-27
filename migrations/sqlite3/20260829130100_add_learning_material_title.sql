-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_learning_materials ADD COLUMN title TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_learning_materials DROP COLUMN title;
-- +goose StatementEnd
