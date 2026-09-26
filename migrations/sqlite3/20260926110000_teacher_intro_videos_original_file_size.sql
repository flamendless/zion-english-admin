-- +goose Up
ALTER TABLE tbl_teacher_intro_videos ADD COLUMN original_file_size INTEGER;

UPDATE tbl_teacher_intro_videos
SET original_file_size = file_size
WHERE source_type = 'upload'
	AND file_size IS NOT NULL;

-- +goose Down
ALTER TABLE tbl_teacher_intro_videos DROP COLUMN original_file_size;
