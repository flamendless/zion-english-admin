-- name: UpsertTeacherMeetingAccount :exec
INSERT INTO tbl_teacher_meeting_accounts (
    teacher_id, service, external_user_id, access_token, refresh_token, token_expires_at, connected_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))
ON CONFLICT(teacher_id, service) DO UPDATE SET
    external_user_id = excluded.external_user_id,
    access_token = excluded.access_token,
    refresh_token = excluded.refresh_token,
    token_expires_at = excluded.token_expires_at,
    updated_at = datetime('now');

-- name: GetTeacherMeetingAccount :one
SELECT id, teacher_id, service, external_user_id, access_token, refresh_token, token_expires_at, connected_at, updated_at
FROM tbl_teacher_meeting_accounts
WHERE teacher_id = ? AND service = ?;

-- name: DeleteTeacherMeetingAccount :exec
DELETE FROM tbl_teacher_meeting_accounts
WHERE teacher_id = ? AND service = ?;

-- name: HasTeacherMeetingAccount :one
SELECT COUNT(*) AS count
FROM tbl_teacher_meeting_accounts
WHERE teacher_id = ? AND service = ?;
