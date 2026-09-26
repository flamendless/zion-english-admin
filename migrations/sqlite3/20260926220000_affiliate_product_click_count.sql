-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_affiliate_products ADD COLUMN click_count INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_affiliate_products DROP COLUMN click_count;
-- +goose StatementEnd
