-- name: InsertTeacherLeaveRequest :exec
INSERT INTO tbl_teacher_leave_requests (
	teacher_id,
	leave_type,
	start_date,
	end_date,
	reason,
	status
)
VALUES (?, ?, ?, ?, ?, 'submitted');

-- name: GetTeacherLeaveRequestByID :one
SELECT
	id,
	teacher_id,
	leave_type,
	start_date,
	end_date,
	reason,
	status,
	reviewed_by,
	created_at,
	reviewed_at
FROM tbl_teacher_leave_requests
WHERE id = ?;

-- name: GetTeacherLeaveRequests :many
SELECT
	l.id,
	l.teacher_id,
	l.leave_type,
	l.start_date,
	l.end_date,
	l.reason,
	l.status,
	l.reviewed_by,
	l.created_at,
	l.reviewed_at,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.assigned_color AS teacher_assigned_color,
	t.profile_picture AS teacher_profile_picture
FROM tbl_teacher_leave_requests l
INNER JOIN tbl_teachers t ON t.id = l.teacher_id
WHERE l.teacher_id = ?
	AND t.deleted = 0
	AND (? = '' OR l.status = ?)
ORDER BY l.created_at DESC;

-- name: ListTeacherLeaveRequests :many
SELECT
	l.id,
	l.teacher_id,
	l.leave_type,
	l.start_date,
	l.end_date,
	l.reason,
	l.status,
	l.reviewed_by,
	l.created_at,
	l.reviewed_at,
	t.first_name AS teacher_first_name,
	t.middle_name AS teacher_middle_name,
	t.last_name AS teacher_last_name,
	t.assigned_color AS teacher_assigned_color,
	t.profile_picture AS teacher_profile_picture
FROM tbl_teacher_leave_requests l
INNER JOIN tbl_teachers t ON t.id = l.teacher_id
WHERE t.deleted = 0
	AND (? = '' OR l.status = ?)
	AND (? = 0 OR l.teacher_id = ?)
ORDER BY l.created_at DESC;

-- name: UpdateTeacherLeaveRequestStatus :exec
UPDATE tbl_teacher_leave_requests
SET
	status = ?,
	reviewed_by = ?,
	reviewed_at = datetime('now')
WHERE id = ?
	AND status = 'submitted';

-- name: CountTeacherLeaveRequestsByStatus :one
SELECT COUNT(*) AS count
FROM tbl_teacher_leave_requests l
INNER JOIN tbl_teachers t ON t.id = l.teacher_id
WHERE l.status = ?
	AND t.deleted = 0;

-- name: CountTeacherLeaveRequestsByTeacherAndStatus :one
SELECT COUNT(*) AS count
FROM tbl_teacher_leave_requests
WHERE teacher_id = ?
	AND status = ?;