-- name: InsertAnnouncement :one
INSERT INTO tbl_announcements (title, description, level, start_date, end_date, visible_to_all, cta_label, cta_url, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateAnnouncement :exec
UPDATE tbl_announcements
SET title = ?, description = ?, level = ?, start_date = ?, end_date = ?, visible_to_all = ?, cta_label = ?, cta_url = ?, status = ?, updated_at = datetime('now')
WHERE id = ?;

-- name: DeleteAnnouncement :exec
UPDATE tbl_announcements
SET status = 'deleted', updated_at = datetime('now')
WHERE id = ? AND status != 'deleted';

-- name: GetAnnouncementByID :one
SELECT id, title, description, level, start_date, end_date, visible_to_all, cta_label, cta_url, status, created_at, updated_at
FROM tbl_announcements
WHERE id = ?;

-- name: CountAnnouncements :one
SELECT COUNT(*) FROM tbl_announcements;

-- name: GetAnnouncementsPaged :many
SELECT id, title, description, level, start_date, end_date, visible_to_all, cta_label, cta_url, status, created_at, updated_at
FROM tbl_announcements
ORDER BY start_date DESC, id DESC
LIMIT ? OFFSET ?;

-- name: GetActiveAnnouncementsAll :many
SELECT id, title, description, level, start_date, end_date, visible_to_all, cta_label, cta_url, status, created_at, updated_at
FROM tbl_announcements
WHERE date(?) BETWEEN start_date AND end_date
AND status = 'published'
ORDER BY
	CASE level WHEN 'critical' THEN 0 WHEN 'warning' THEN 1 ELSE 2 END,
	start_date ASC,
	id ASC;

-- name: GetActiveAnnouncementsForTeacher :many
SELECT a.id, a.title, a.description, a.level, a.start_date, a.end_date, a.visible_to_all, a.cta_label, a.cta_url, a.status, a.created_at, a.updated_at
FROM tbl_announcements a
WHERE date(?) BETWEEN a.start_date AND a.end_date
AND a.status = 'published'
AND (
	a.visible_to_all = 1
	OR EXISTS (
		SELECT 1 FROM tbl_announcements_teachers_m2m m
		WHERE m.announcement_id = a.id AND m.teacher_id = ?
	)
)
ORDER BY
	CASE a.level WHEN 'critical' THEN 0 WHEN 'warning' THEN 1 ELSE 2 END,
	a.start_date ASC,
	a.id ASC;

-- name: GetTeacherIDsByAnnouncementID :many
SELECT teacher_id FROM tbl_announcements_teachers_m2m WHERE announcement_id = ?;

-- name: DeleteAnnouncementTeacherLinks :exec
DELETE FROM tbl_announcements_teachers_m2m WHERE announcement_id = ?;

-- name: InsertAnnouncementTeacherM2M :exec
INSERT INTO tbl_announcements_teachers_m2m (announcement_id, teacher_id) VALUES (?, ?);

-- name: CountTeachersByAnnouncementID :one
SELECT COUNT(*) FROM tbl_announcements_teachers_m2m WHERE announcement_id = ?;
