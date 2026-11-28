-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_teacher_payments ADD COLUMN receipt_nudge INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_teacher_payments DROP COLUMN receipt_nudge;
-- +goose StatementEnd