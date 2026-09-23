-- name: InsertTeacherPayment :exec
INSERT INTO tbl_teacher_payments (
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
	sent_at,
	created_at,
	updated_at
) VALUES (
	?,
	?,
	?,
	?,
	?,
	?,
	?,
	?,
	?,
	'pending',
	datetime('now'),
	datetime('now'),
	datetime('now')
);

-- name: GetTeacherPaymentsFiltered :many
SELECT
	p.id,
	p.teacher_id,
	p.sent_by_teacher_id,
	p.sent_by_name,
	p.payment_method,
	p.reference_number,
	p.amount,
	p.currency,
	p.period_start,
	p.period_end,
	p.status,
	p.sent_at,
	p.received_at,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.profile_picture AS teacher_profile_picture
FROM tbl_teacher_payments p
INNER JOIN tbl_teachers t ON t.id = p.teacher_id
WHERE t.deleted = 0
	AND (? = '' OR p.period_start >= ?)
	AND (? = '' OR p.period_end <= ?)
	AND (
		? = ''
		OR t.first_name LIKE '%' || ? || '%'
		OR t.middle_name LIKE '%' || ? || '%'
		OR t.last_name LIKE '%' || ? || '%'
	)
	AND (? = '' OR p.status = ?)
	AND (
		? = 0
		OR p.teacher_id = ?
	)
ORDER BY p.sent_at DESC;

-- name: TeacherHasPaymentForPeriod :one
SELECT EXISTS (
	SELECT 1
	FROM tbl_teacher_payments
	WHERE teacher_id = ?
		AND period_start = ?
		AND period_end = ?
) AS has_payment;

-- name: GetTeacherPaymentStatusesForPeriod :many
SELECT
	teacher_id,
	status
FROM tbl_teacher_payments
WHERE period_start = ?
	AND period_end = ?
ORDER BY
	CASE status WHEN 'pending' THEN 0 ELSE 1 END,
	sent_at DESC;

-- name: GetPendingPaymentForTeacherReceipt :one
SELECT
	p.id,
	p.teacher_id,
	p.payment_method,
	p.reference_number,
	p.amount,
	p.currency,
	p.period_start,
	p.period_end,
	p.sent_at,
	p.dismissed_access_id
FROM tbl_teacher_payments p
WHERE p.teacher_id = ?
	AND p.status = 'pending'
	AND (
		p.dismissed_access_id IS NULL
		OR (
			? > 0
			AND p.dismissed_access_id != ?
		)
	)
ORDER BY p.sent_at ASC
LIMIT 1;

-- name: GetTeacherPaymentByID :one
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
	received_at
FROM tbl_teacher_payments
WHERE id = ?
LIMIT 1;

-- name: MarkTeacherPaymentReceived :exec
UPDATE tbl_teacher_payments
SET
	status = 'received',
	received_at = datetime('now'),
	updated_at = datetime('now')
WHERE id = ?
	AND teacher_id = ?
	AND status = 'pending';

-- name: DeferTeacherPaymentReceipt :exec
UPDATE tbl_teacher_payments
SET
	dismissed_access_id = ?,
	updated_at = datetime('now')
WHERE id = ?
	AND teacher_id = ?
	AND status = 'pending';
