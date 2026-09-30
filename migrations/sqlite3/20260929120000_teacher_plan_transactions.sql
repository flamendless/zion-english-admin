-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_teacher_plan_transactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	tier TEXT NOT NULL CHECK (tier IN ('pro')),
	billing_kind TEXT NOT NULL CHECK (billing_kind IN ('monthly', 'lifetime')),
	effective_start TEXT NOT NULL,
	effective_end TEXT NULL,
	revoked_at TEXT NULL,
	granted_by_teacher_id INTEGER NULL REFERENCES tbl_teachers(id),
	granted_by_name TEXT NOT NULL,
	note TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_teacher_plan_txn_teacher_start ON tbl_teacher_plan_transactions(teacher_id, effective_start DESC);
CREATE INDEX IF NOT EXISTS idx_teacher_plan_txn_teacher_revoked ON tbl_teacher_plan_transactions(teacher_id, revoked_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_teacher_plan_txn_teacher_revoked;
DROP INDEX IF EXISTS idx_teacher_plan_txn_teacher_start;
DROP TABLE IF EXISTS tbl_teacher_plan_transactions;
-- +goose StatementEnd
