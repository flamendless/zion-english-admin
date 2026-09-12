-- name: InsertClassCalendarEvent :one
INSERT INTO tbl_class_calendar_event (class_id, service, event_id, event_url)
VALUES (?, ?, ?, ?)
RETURNING id, class_id, service, event_id, event_url, created_at, updated_at, deleted_at;

-- name: GetActiveClassCalendarEventByClassID :one
SELECT id, class_id, service, event_id, event_url, created_at, updated_at, deleted_at
FROM tbl_class_calendar_event
WHERE class_id = ? AND deleted_at IS NULL
ORDER BY id DESC
LIMIT 1;

-- name: GetActiveClassCalendarEventsByClassIDs :many
SELECT id, class_id, service, event_id, event_url, created_at, updated_at, deleted_at
FROM tbl_class_calendar_event
WHERE class_id IN (sqlc.slice('class_ids')) AND deleted_at IS NULL;

-- name: UpdateClassCalendarEvent :exec
UPDATE tbl_class_calendar_event
SET event_url = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: SoftDeleteClassCalendarEvent :exec
UPDATE tbl_class_calendar_event
SET deleted_at = datetime('now'), updated_at = datetime('now')
WHERE id = ?;
