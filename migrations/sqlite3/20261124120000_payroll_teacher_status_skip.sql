-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_payroll_session_teachers_skip_status (
	session_id INTEGER NOT NULL REFERENCES tbl_payroll_sessions(id) ON DELETE CASCADE,
	teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	status TEXT NOT NULL DEFAULT 'accomplish_daily_logs' CHECK (
		status IN (
			'accomplish_daily_logs',
			'done_daily_logs',
			'audited',
			'sent_salary',
			'salary_received',
			'skip'
		)
	),
	payment_id INTEGER NULL REFERENCES tbl_teacher_payments(id),
	daily_logs_dismissed_access_id INTEGER NULL REFERENCES tbl_accesses(id),
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	updated_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (session_id, teacher_id)
);

INSERT INTO tbl_payroll_session_teachers_skip_status (
	session_id,
	teacher_id,
	status,
	payment_id,
	daily_logs_dismissed_access_id,
	created_at,
	updated_at
)
SELECT
	session_id,
	teacher_id,
	status,
	payment_id,
	daily_logs_dismissed_access_id,
	created_at,
	updated_at
FROM tbl_payroll_session_teachers;

DROP TABLE tbl_payroll_session_teachers;
ALTER TABLE tbl_payroll_session_teachers_skip_status RENAME TO tbl_payroll_session_teachers;

CREATE INDEX IF NOT EXISTS idx_payroll_session_teachers_teacher ON tbl_payroll_session_teachers(teacher_id);

CREATE TABLE tbl_payroll_session_teacher_status_events_skip_status (
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
			'salary_received',
			'skip'
		)
	),
	reason TEXT NULL,
	changed_by_teacher_id INTEGER NOT NULL REFERENCES tbl_teachers(id),
	changed_by_name TEXT NOT NULL DEFAULT '',
	changed_at TEXT NOT NULL DEFAULT (datetime('now')),
	FOREIGN KEY (session_id, teacher_id) REFERENCES tbl_payroll_session_teachers(session_id, teacher_id) ON DELETE CASCADE
);

INSERT INTO tbl_payroll_session_teacher_status_events_skip_status (
	id,
	session_id,
	teacher_id,
	from_status,
	to_status,
	reason,
	changed_by_teacher_id,
	changed_by_name,
	changed_at
)
SELECT
	id,
	session_id,
	teacher_id,
	from_status,
	to_status,
	reason,
	changed_by_teacher_id,
	changed_by_name,
	changed_at
FROM tbl_payroll_session_teacher_status_events;

DROP TABLE tbl_payroll_session_teacher_status_events;
ALTER TABLE tbl_payroll_session_teacher_status_events_skip_status RENAME TO tbl_payroll_session_teacher_status_events;

CREATE INDEX IF NOT EXISTS idx_payroll_teacher_status_events_lookup
	ON tbl_payroll_session_teacher_status_events(session_id, teacher_id, changed_at ASC);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_payroll_session_teachers_legacy_status (
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

INSERT INTO tbl_payroll_session_teachers_legacy_status (
	session_id,
	teacher_id,
	status,
	payment_id,
	daily_logs_dismissed_access_id,
	created_at,
	updated_at
)
SELECT
	session_id,
	teacher_id,
	status,
	payment_id,
	daily_logs_dismissed_access_id,
	created_at,
	updated_at
FROM tbl_payroll_session_teachers
WHERE status != 'skip';

DROP TABLE tbl_payroll_session_teachers;
ALTER TABLE tbl_payroll_session_teachers_legacy_status RENAME TO tbl_payroll_session_teachers;

CREATE INDEX IF NOT EXISTS idx_payroll_session_teachers_teacher ON tbl_payroll_session_teachers(teacher_id);

CREATE TABLE tbl_payroll_session_teacher_status_events_legacy_status (
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
	changed_by_name TEXT NOT NULL DEFAULT '',
	changed_at TEXT NOT NULL DEFAULT (datetime('now')),
	FOREIGN KEY (session_id, teacher_id) REFERENCES tbl_payroll_session_teachers(session_id, teacher_id) ON DELETE CASCADE
);

INSERT INTO tbl_payroll_session_teacher_status_events_legacy_status (
	id,
	session_id,
	teacher_id,
	from_status,
	to_status,
	reason,
	changed_by_teacher_id,
	changed_by_name,
	changed_at
)
SELECT
	id,
	session_id,
	teacher_id,
	from_status,
	to_status,
	reason,
	changed_by_teacher_id,
	changed_by_name,
	changed_at
FROM tbl_payroll_session_teacher_status_events
WHERE to_status != 'skip';

DROP TABLE tbl_payroll_session_teacher_status_events;
ALTER TABLE tbl_payroll_session_teacher_status_events_legacy_status RENAME TO tbl_payroll_session_teacher_status_events;

CREATE INDEX IF NOT EXISTS idx_payroll_teacher_status_events_lookup
	ON tbl_payroll_session_teacher_status_events(session_id, teacher_id, changed_at ASC);

PRAGMA foreign_keys=ON;
-- +goose StatementEnd