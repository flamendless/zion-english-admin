-- +goose Up
-- +goose StatementBegin
UPDATE tbl_ads
SET placement = 'top_and_bottom'
WHERE placement = 'top_and_side';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE tbl_ads
SET placement = 'top_and_side'
WHERE placement = 'top_and_bottom';
-- +goose StatementEnd
