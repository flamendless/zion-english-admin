-- +goose Up
-- Idempotent safety net when 20261002120000 applied ALTERs/dedupe but index step was skipped.
-- No-op when 20261002120000 completed fully (index already exists).

-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS idx_affiliate_products_provider_item_id
	ON tbl_affiliate_products (provider, item_id)
	WHERE trim(item_id) != '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 1;
-- +goose StatementEnd
