-- +goose Up
DROP INDEX IF EXISTS idx_teacher_intro_videos_blocking;
CREATE UNIQUE INDEX IF NOT EXISTS idx_teacher_intro_videos_blocking
ON tbl_teacher_intro_videos(teacher_id)
WHERE status IN ('processing', 'submitted', 'approved');

-- +goose Down
DROP INDEX IF EXISTS idx_teacher_intro_videos_blocking;
CREATE UNIQUE INDEX IF NOT EXISTS idx_teacher_intro_videos_blocking
ON tbl_teacher_intro_videos(teacher_id)
WHERE status IN ('submitted', 'approved');
