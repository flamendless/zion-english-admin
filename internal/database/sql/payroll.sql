-- name: InsertPayrollSession :one
INSERT INTO tbl_payroll_sessions (
	period_start,
	period_end,
	role_teacher,
	role_admin,
	status,
	started_by_teacher_id,
	started_at,
	created_at,
	updated_at
) VALUES (
	?,
	?,
	?,
	?,
	'active',
	?,
	datetime('now'),
	datetime('now'),
	datetime('now')
)
RETURNING id;

-- name: GetPayrollSessionByID :one
SELECT
	id,
	period_start,
	period_end,
	role_teacher,
	role_admin,
	status,
	started_by_teacher_id,
	started_at,
	finished_at,
	created_at,
	updated_at
FROM tbl_payroll_sessions
WHERE id = ?;

-- name: GetActivePayrollSession :one
SELECT
	id,
	period_start,
	period_end,
	role_teacher,
	role_admin,
	status,
	started_by_teacher_id,
	started_at,
	finished_at,
	created_at,
	updated_at
FROM tbl_payroll_sessions
WHERE status = 'active'
LIMIT 1;

-- name: CountPayrollSessions :one
SELECT COUNT(*) FROM tbl_payroll_sessions;

-- name: ListPayrollSessions :many
SELECT
	s.id,
	s.period_start,
	s.period_end,
	s.role_teacher,
	s.role_admin,
	s.status,
	s.started_by_teacher_id,
	s.started_at,
	s.finished_at,
	st.first_name AS started_by_first_name,
	st.middle_name AS started_by_middle_name,
	st.last_name AS started_by_last_name,
	(
		SELECT COUNT(*)
		FROM tbl_payroll_session_teachers pst
		WHERE pst.session_id = s.id
	) AS teacher_count,
	(
		SELECT COUNT(*)
		FROM tbl_payroll_session_teachers pst
		WHERE pst.session_id = s.id
			AND pst.status = 'salary_received'
	) AS salary_received_count
FROM tbl_payroll_sessions s
INNER JOIN tbl_teachers st ON st.id = s.started_by_teacher_id
ORDER BY s.started_at DESC
LIMIT ?
OFFSET ?;

-- name: FinishPayrollSession :exec
UPDATE tbl_payroll_sessions
SET
	status = 'finished',
	finished_at = datetime('now'),
	updated_at = datetime('now')
WHERE id = ?
	AND status = 'active';

-- name: InsertPayrollSessionTeacher :exec
INSERT INTO tbl_payroll_session_teachers (
	session_id,
	teacher_id,
	status,
	created_at,
	updated_at
) VALUES (
	?,
	?,
	'accomplish_daily_logs',
	datetime('now'),
	datetime('now')
);

-- name: ListPayrollSessionTeachers :many
SELECT
	pst.session_id,
	pst.teacher_id,
	pst.status,
	pst.payment_id,
	pst.daily_logs_dismissed_access_id,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.profile_picture AS teacher_profile_picture,
	t.status AS teacher_account_status
FROM tbl_payroll_session_teachers pst
INNER JOIN tbl_teachers t ON t.id = pst.teacher_id
WHERE pst.session_id = ?
	AND t.deleted = 0
ORDER BY t.last_name ASC, t.first_name ASC;

-- name: GetPayrollSessionTeacher :one
SELECT
	pst.session_id,
	pst.teacher_id,
	pst.status,
	pst.payment_id,
	pst.daily_logs_dismissed_access_id,
	s.period_start,
	s.period_end,
	s.status AS session_status
FROM tbl_payroll_session_teachers pst
INNER JOIN tbl_payroll_sessions s ON s.id = pst.session_id
WHERE pst.session_id = ?
	AND pst.teacher_id = ?;

-- name: UpdatePayrollSessionTeacherStatus :exec
UPDATE tbl_payroll_session_teachers
SET
	status = ?,
	updated_at = datetime('now')
WHERE session_id = ?
	AND teacher_id = ?
	AND status = ?;

-- name: SetPayrollSessionTeacherPayment :exec
UPDATE tbl_payroll_session_teachers
SET
	status = 'sent_salary',
	payment_id = ?,
	updated_at = datetime('now')
WHERE session_id = ?
	AND teacher_id = ?
	AND status = 'audited';

-- name: SetPayrollSessionTeacherSalaryReceived :exec
UPDATE tbl_payroll_session_teachers
SET
	status = 'salary_received',
	updated_at = datetime('now')
WHERE session_id = ?
	AND teacher_id = ?
	AND payment_id = ?
	AND status = 'sent_salary';

-- name: CountPayrollSessionTeachersNotSalaryReceived :one
SELECT COUNT(*)
FROM tbl_payroll_session_teachers
WHERE session_id = ?
	AND status != 'salary_received';

-- name: GetActivePayrollSessionTeacherRow :one
SELECT
	pst.session_id,
	pst.teacher_id,
	pst.status,
	pst.daily_logs_dismissed_access_id,
	s.period_start,
	s.period_end,
	s.started_at
FROM tbl_payroll_session_teachers pst
INNER JOIN tbl_payroll_sessions s ON s.id = pst.session_id
WHERE s.status = 'active'
	AND pst.teacher_id = ?;

-- name: DeferPayrollDailyLogs :exec
UPDATE tbl_payroll_session_teachers
SET
	daily_logs_dismissed_access_id = ?,
	updated_at = datetime('now')
WHERE session_id = ?
	AND teacher_id = ?
	AND status = 'accomplish_daily_logs';

-- name: GetPayrollSessionTeacherForDailyLogsPrompt :one
SELECT
	pst.session_id,
	pst.teacher_id,
	pst.status,
	pst.daily_logs_dismissed_access_id,
	s.period_start,
	s.period_end
FROM tbl_payroll_session_teachers pst
INNER JOIN tbl_payroll_sessions s ON s.id = pst.session_id
WHERE s.status = 'active'
	AND pst.teacher_id = ?
	AND pst.status = 'accomplish_daily_logs'
	AND (
		pst.daily_logs_dismissed_access_id IS NULL
		OR pst.daily_logs_dismissed_access_id != ?
	);

-- name: ConfirmPayrollDailyLogs :exec
UPDATE tbl_payroll_session_teachers
SET
	status = 'done_daily_logs',
	updated_at = datetime('now')
WHERE session_id = ?
	AND teacher_id = ?
	AND status = 'accomplish_daily_logs';

-- name: GetPayrollSessionTeacherByPaymentID :one
SELECT
	pst.session_id,
	pst.teacher_id,
	pst.status,
	pst.payment_id,
	s.period_start,
	s.period_end,
	s.status AS session_status
FROM tbl_payroll_session_teachers pst
INNER JOIN tbl_payroll_sessions s ON s.id = pst.session_id
WHERE pst.payment_id = ?
	AND s.status = 'active';

-- name: GetActivePayrollSessionTeacherForPeriod :one
SELECT
	pst.session_id,
	pst.teacher_id,
	pst.status,
	pst.payment_id
FROM tbl_payroll_session_teachers pst
INNER JOIN tbl_payroll_sessions s ON s.id = pst.session_id
WHERE s.status = 'active'
	AND pst.teacher_id = ?
	AND s.period_start = ?
	AND s.period_end = ?;

-- name: InsertPayrollSessionTeacherStatusEvent :exec
INSERT INTO tbl_payroll_session_teacher_status_events (
	session_id,
	teacher_id,
	from_status,
	to_status,
	reason,
	changed_by_teacher_id,
	changed_by_name,
	changed_at
) VALUES (
	?,
	?,
	?,
	?,
	?,
	?,
	?,
	datetime('now')
);

-- name: ListPayrollSessionTeacherStatusEvents :many
SELECT
	e.id,
	e.from_status,
	e.to_status,
	e.reason,
	e.changed_at,
	e.changed_by_teacher_id,
	e.changed_by_name,
	t.first_name AS changed_by_first_name,
	t.middle_name AS changed_by_middle_name,
	t.last_name AS changed_by_last_name
FROM tbl_payroll_session_teacher_status_events e
LEFT JOIN tbl_teachers t ON t.id = e.changed_by_teacher_id
WHERE e.session_id = ?
	AND e.teacher_id = ?
ORDER BY e.changed_at ASC, e.id ASC;

-- name: SetPayrollSessionTeacherStatusDirect :exec
UPDATE tbl_payroll_session_teachers
SET
	status = ?,
	updated_at = datetime('now')
WHERE session_id = ?
	AND teacher_id = ?;

-- name: ClearPayrollSessionTeacherPayment :exec
UPDATE tbl_payroll_session_teachers
SET
	payment_id = NULL,
	updated_at = datetime('now')
WHERE session_id = ?
	AND teacher_id = ?;
