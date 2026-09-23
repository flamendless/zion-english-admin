-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_teacher_payments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	sent_by_teacher_id INTEGER NULL REFERENCES tbl_teachers(id),
	sent_by_name TEXT NOT NULL,
	payment_method TEXT NOT NULL CHECK (payment_method IN ('gcash')),
	reference_number TEXT NOT NULL,
	amount REAL NOT NULL CHECK (amount > 0),
	currency TEXT NOT NULL CHECK (currency IN ('KRW', 'CAD', 'YEN', 'PHP')),
	period_start TEXT NOT NULL,
	period_end TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'received')),
	dismissed_access_id INTEGER NULL REFERENCES tbl_accesses(id),
	sent_at TEXT NOT NULL DEFAULT (datetime('now')),
	received_at TEXT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_teacher_payments_teacher_status ON tbl_teacher_payments(teacher_id, status);
CREATE INDEX IF NOT EXISTS idx_teacher_payments_period ON tbl_teacher_payments(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_teacher_payments_sent_at ON tbl_teacher_payments(sent_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_teacher_payments_sent_at;
DROP INDEX IF EXISTS idx_teacher_payments_period;
DROP INDEX IF EXISTS idx_teacher_payments_teacher_status;
DROP TABLE IF EXISTS tbl_teacher_payments;
-- +goose StatementEnd
