-- name: InsertTeacherPlanTransaction :one
INSERT INTO tbl_teacher_plan_transactions (
	teacher_id,
	tier,
	billing_kind,
	effective_start,
	effective_end,
	granted_by_teacher_id,
	granted_by_name,
	note
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: RevokeTeacherPlanTransaction :execrows
UPDATE tbl_teacher_plan_transactions
SET
	revoked_at = datetime('now'),
	effective_end = CASE
		WHEN effective_end IS NULL OR effective_end > datetime('now') THEN datetime('now')
		ELSE effective_end
	END
WHERE id = ?
	AND teacher_id = ?
	AND revoked_at IS NULL;

-- name: GetTeacherPlanTransactionsByTeacherID :many
SELECT *
FROM tbl_teacher_plan_transactions
WHERE teacher_id = ?
ORDER BY effective_start DESC, id DESC
LIMIT ?
OFFSET ?;

-- name: CountTeacherPlanTransactionsByTeacherID :one
SELECT COUNT(*) AS count
FROM tbl_teacher_plan_transactions
WHERE teacher_id = ?;

-- name: GetEffectiveTeacherPlanGrant :one
SELECT *
FROM tbl_teacher_plan_transactions
WHERE teacher_id = ?
	AND tier = 'pro'
	AND revoked_at IS NULL
	AND effective_start <= datetime('now')
	AND (
		effective_end IS NULL
		OR effective_end > datetime('now')
	)
ORDER BY effective_start DESC, id DESC
LIMIT 1;

-- name: ListRecentTeacherPlanTransactions :many
SELECT *
FROM tbl_teacher_plan_transactions
ORDER BY created_at DESC, id DESC
LIMIT ?
OFFSET ?;

-- name: CountAllTeacherPlanTransactions :one
SELECT COUNT(*) AS count
FROM tbl_teacher_plan_transactions;
