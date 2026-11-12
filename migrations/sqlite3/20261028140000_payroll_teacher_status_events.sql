-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_payroll_session_teacher_status_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id INTEGER NOT NULL REFERENCES tbl_payroll_sessions(id) ON DELETE CASCADE,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	from_status TEXT NULL,
	to_status TEXT NOT NULL CHECK (
		to_status IN (
			'accomplish_daily_logs',
			'done_daily_logs',
			'audited',
			'sent_salary',
			'salary_received'
		)
	),
	reason TEXT NULL,
	changed_by_teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	changed_at TEXT NOT NULL DEFAULT (datetime('now')),
	FOREIGN KEY (session_id, teacher_id) REFERENCES tbl_payroll_session_teachers(session_id, teacher_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_payroll_teacher_status_events_lookup
	ON tbl_payroll_session_teacher_status_events(session_id, teacher_id, changed_at ASC);

INSERT INTO tbl_payroll_session_teacher_status_events (
	session_id,
	teacher_id,
	from_status,
	to_status,
	reason,
	changed_by_teacher_id,
	changed_at
)
SELECT
	pst.session_id,
	pst.teacher_id,
	NULL,
	pst.status,
	NULL,
	s.started_by_teacher_id,
	pst.created_at
FROM tbl_payroll_session_teachers pst
INNER JOIN tbl_payroll_sessions s ON s.id = pst.session_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_payroll_teacher_status_events_lookup;
DROP TABLE IF EXISTS tbl_payroll_session_teacher_status_events;
-- +goose StatementEnd
