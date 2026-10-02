-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_affiliate_products ADD COLUMN thumbnail_orientation TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_affiliate_products DROP COLUMN thumbnail_orientation;
-- +goose StatementEnd
