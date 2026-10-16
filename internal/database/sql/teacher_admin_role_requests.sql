-- name: CountTeacherAdminRoleRequestsByStatus :one
SELECT COUNT(*) AS count
FROM tbl_teacher_admin_role_requests r
INNER JOIN tbl_teachers target ON target.id = r.target_teacher_id
WHERE r.status = ?
	AND target.deleted = 0;

-- name: GetTeacherAdminRoleRequestsByStatus :many
SELECT
	r.id,
	r.target_teacher_id,
	r.requested_by_teacher_id,
	r.status,
	r.reviewed_by,
	r.created_at,
	r.reviewed_at,
	target.first_name AS target_first_name,
	target.middle_name AS target_middle_name,
	target.last_name AS target_last_name,
	target.assigned_color AS target_assigned_color,
	target.profile_picture AS target_profile_picture,
	requester.first_name AS requester_first_name,
	requester.middle_name AS requester_middle_name,
	requester.last_name AS requester_last_name,
	requester.assigned_color AS requester_assigned_color,
	requester.profile_picture AS requester_profile_picture
FROM tbl_teacher_admin_role_requests r
INNER JOIN tbl_teachers target ON target.id = r.target_teacher_id
INNER JOIN tbl_teachers requester ON requester.id = r.requested_by_teacher_id
WHERE r.status = ?
	AND target.deleted = 0
ORDER BY r.created_at DESC;

-- name: GetPendingTeacherAdminRoleRequestByTargetTeacherID :one
SELECT
	id,
	target_teacher_id,
	requested_by_teacher_id,
	status,
	reviewed_by,
	created_at,
	reviewed_at
FROM tbl_teacher_admin_role_requests
WHERE target_teacher_id = ?
	AND status = 'pending';

-- name: GetTeacherAdminRoleRequestByID :one
SELECT
	id,
	target_teacher_id,
	requested_by_teacher_id,
	status,
	reviewed_by,
	created_at,
	reviewed_at
FROM tbl_teacher_admin_role_requests
WHERE id = ?;

-- name: InsertTeacherAdminRoleRequest :exec
INSERT INTO tbl_teacher_admin_role_requests (target_teacher_id, requested_by_teacher_id, status)
VALUES (?, ?, 'pending');

-- name: DeletePendingTeacherAdminRoleRequestByTarget :exec
DELETE FROM tbl_teacher_admin_role_requests
WHERE target_teacher_id = ?
	AND status = 'pending';

-- name: UpdateTeacherAdminRoleRequestStatus :exec
UPDATE tbl_teacher_admin_role_requests
SET
	status = ?,
	reviewed_by = ?,
	reviewed_at = datetime('now')
WHERE id = ?
	AND status = 'pending';
