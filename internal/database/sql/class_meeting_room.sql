-- name: InsertClassMeetingRoom :one
INSERT INTO tbl_class_meeting_room (class_id, service, room_id, room_url, room_passcode)
VALUES (?, ?, ?, ?, ?)
RETURNING id, class_id, service, room_id, room_url, room_passcode, created_at, updated_at, deleted_at;

-- name: GetActiveClassMeetingRoomByClassID :one
SELECT id, class_id, service, room_id, room_url, room_passcode, created_at, updated_at, deleted_at
FROM tbl_class_meeting_room
WHERE class_id = ? AND deleted_at IS NULL
ORDER BY id DESC
LIMIT 1;

-- name: GetActiveClassMeetingRoomsByClassIDs :many
SELECT id, class_id, service, room_id, room_url, room_passcode, created_at, updated_at, deleted_at
FROM tbl_class_meeting_room
WHERE class_id IN (sqlc.slice('class_ids')) AND deleted_at IS NULL;

-- name: UpdateClassMeetingRoom :exec
UPDATE tbl_class_meeting_room
SET room_url = ?, room_passcode = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: SoftDeleteClassMeetingRoom :exec
UPDATE tbl_class_meeting_room
SET deleted_at = datetime('now'), updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteScheduledClassByID :exec
DELETE FROM tbl_scheduled_classes
WHERE id = ?;
