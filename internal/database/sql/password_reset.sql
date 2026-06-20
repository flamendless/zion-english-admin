-- name: InsertPasswordResetEvent :exec
INSERT INTO tbl_password_reset_events (email, ip_address, teacher_id, reset_token, status, event, expires_at)
VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: CountPasswordResetRequestsByIPSince :one
SELECT COUNT(*) as count
FROM tbl_password_reset_events
WHERE ip_address = ?
  AND event = 'request_submitted'
  AND created_at > ?;

-- name: GetPasswordResetByToken :one
SELECT id, email, ip_address, teacher_id, reset_token, status, event, created_at, expires_at
FROM tbl_password_reset_events
WHERE reset_token = ?
  AND status = 'token_issued'
ORDER BY created_at DESC
LIMIT 1;

-- name: HasCompletedPasswordResetForToken :one
SELECT COUNT(*) as count
FROM tbl_password_reset_events
WHERE reset_token = ?
  AND status = 'completed';
