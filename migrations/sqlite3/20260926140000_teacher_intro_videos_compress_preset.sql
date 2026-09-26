-- +goose Up
ALTER TABLE tbl_teacher_intro_videos ADD COLUMN compress_preset TEXT CHECK(
	compress_preset IS NULL
	OR compress_preset IN ('very_low', 'medium', 'fine', 'high', 'original')
);

-- +goose Down
ALTER TABLE tbl_teacher_intro_videos DROP COLUMN compress_preset;
