-- +goose Up
-- +goose StatementBegin
ALTER TABLE tbl_payroll_session_teacher_status_events
	ADD COLUMN changed_by_name TEXT NOT NULL DEFAULT '';

UPDATE tbl_payroll_session_teacher_status_events
SET changed_by_name = trim(
	coalesce(
		(
			SELECT t.first_name || ' ' || t.last_name
			FROM tbl_teachers t
			WHERE t.id = tbl_payroll_session_teacher_status_events.changed_by_teacher_id
		),
		''
	)
)
WHERE changed_by_name = '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tbl_payroll_session_teacher_status_events DROP COLUMN changed_by_name;
-- +goose StatementEnd
