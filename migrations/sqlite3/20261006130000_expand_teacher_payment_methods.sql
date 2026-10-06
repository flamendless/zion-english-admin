-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teacher_payments_expanded_methods (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	sent_by_teacher_id INTEGER NULL REFERENCES tbl_teachers(id),
	sent_by_name TEXT NOT NULL,
	payment_method TEXT NOT NULL CHECK (payment_method IN (
		'gcash', 'bpi', 'gotyme', 'pnb', 'unionbank', 'bdo', 'maya'
	)),
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

INSERT INTO tbl_teacher_payments_expanded_methods (
	id,
	teacher_id,
	sent_by_teacher_id,
	sent_by_name,
	payment_method,
	reference_number,
	amount,
	currency,
	period_start,
	period_end,
	status,
	dismissed_access_id,
	sent_at,
	received_at,
	created_at,
	updated_at
)
SELECT
	id,
	teacher_id,
	sent_by_teacher_id,
	sent_by_name,
	payment_method,
	reference_number,
	amount,
	currency,
	period_start,
	period_end,
	status,
	dismissed_access_id,
	sent_at,
	received_at,
	created_at,
	updated_at
FROM tbl_teacher_payments;

DROP TABLE tbl_teacher_payments;
ALTER TABLE tbl_teacher_payments_expanded_methods RENAME TO tbl_teacher_payments;

CREATE INDEX IF NOT EXISTS idx_teacher_payments_teacher_status ON tbl_teacher_payments(teacher_id, status);
CREATE INDEX IF NOT EXISTS idx_teacher_payments_period ON tbl_teacher_payments(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_teacher_payments_sent_at ON tbl_teacher_payments(sent_at DESC);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_teacher_payments_gcash_only (
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

INSERT INTO tbl_teacher_payments_gcash_only (
	id,
	teacher_id,
	sent_by_teacher_id,
	sent_by_name,
	payment_method,
	reference_number,
	amount,
	currency,
	period_start,
	period_end,
	status,
	dismissed_access_id,
	sent_at,
	received_at,
	created_at,
	updated_at
)
SELECT
	id,
	teacher_id,
	sent_by_teacher_id,
	sent_by_name,
	payment_method,
	reference_number,
	amount,
	currency,
	period_start,
	period_end,
	status,
	dismissed_access_id,
	sent_at,
	received_at,
	created_at,
	updated_at
FROM tbl_teacher_payments
WHERE payment_method = 'gcash';

DROP TABLE tbl_teacher_payments;
ALTER TABLE tbl_teacher_payments_gcash_only RENAME TO tbl_teacher_payments;

CREATE INDEX IF NOT EXISTS idx_teacher_payments_teacher_status ON tbl_teacher_payments(teacher_id, status);
CREATE INDEX IF NOT EXISTS idx_teacher_payments_period ON tbl_teacher_payments(period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_teacher_payments_sent_at ON tbl_teacher_payments(sent_at DESC);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
