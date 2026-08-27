-- +goose Up
ALTER TABLE tbl_learning_materials ADD COLUMN thumbnail_url TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE tbl_learning_materials DROP COLUMN thumbnail_url;
