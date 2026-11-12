-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_payroll_sessions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	period_start TEXT NOT NULL,
	period_end TEXT NOT NULL,
	role_teacher INTEGER NOT NULL DEFAULT 1 CHECK (role_teacher IN (0, 1)),
	role_admin INTEGER NOT NULL DEFAULT 0 CHECK (role_admin IN (0, 1)),
	status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'finished')),
	started_by_teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	started_at TEXT NOT NULL DEFAULT (datetime('now')),
	finished_at TEXT NULL,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_payroll_sessions_one_active
	ON tbl_payroll_sessions(status)
	WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_payroll_sessions_started_at ON tbl_payroll_sessions(started_at DESC);

CREATE TABLE IF NOT EXISTS tbl_payroll_session_teachers (
	session_id INTEGER NOT NULL REFERENCES tbl_payroll_sessions(id) ON DELETE CASCADE,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	status TEXT NOT NULL DEFAULT 'accomplish_daily_logs' CHECK (
		status IN (
			'accomplish_daily_logs',
			'done_daily_logs',
			'audited',
			'sent_salary',
			'salary_received'
		)
	),
	payment_id INTEGER NULL REFERENCES tbl_teacher_payments(id),
	daily_logs_dismissed_access_id INTEGER NULL REFERENCES tbl_accesses(id),
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (session_id, teacher_id)
);

CREATE INDEX IF NOT EXISTS idx_payroll_session_teachers_teacher ON tbl_payroll_session_teachers(teacher_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payroll_session_teachers_teacher;
DROP TABLE IF EXISTS tbl_payroll_session_teachers;
DROP INDEX IF EXISTS idx_payroll_sessions_started_at;
DROP INDEX IF EXISTS idx_payroll_sessions_one_active;
DROP TABLE IF EXISTS tbl_payroll_sessions;
-- +goose StatementEnd
