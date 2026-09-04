-- +goose Up
-- +goose StatementBegin
INSERT INTO tbl_students (name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency, assigned_color, status, inactive_reason)
SELECT 'Test Student Alpha', 'PHP', NULL, 100, NULL, NULL, NULL, '#90C020', 'active', NULL
WHERE NOT EXISTS (SELECT 1 FROM tbl_students WHERE name = 'Test Student Alpha');

INSERT INTO tbl_students (name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency, assigned_color, status, inactive_reason)
SELECT 'Test Student Beta', 'PHP', NULL, 100, NULL, NULL, NULL, '#90C020', 'active', NULL
WHERE NOT EXISTS (SELECT 1 FROM tbl_students WHERE name = 'Test Student Beta');

INSERT INTO tbl_students (name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency, assigned_color, status, inactive_reason)
SELECT 'Test Student Gamma', 'PHP', NULL, 100, NULL, NULL, NULL, '#90C020', 'active', NULL
WHERE NOT EXISTS (SELECT 1 FROM tbl_students WHERE name = 'Test Student Gamma');

INSERT INTO tbl_students (name, currency, contact, rate_per_class, parent_name, parent_rate, parent_currency, assigned_color, status, inactive_reason)
SELECT 'Test Student Delta', 'PHP', NULL, 100, NULL, NULL, NULL, '#90C020', 'active', NULL
WHERE NOT EXISTS (SELECT 1 FROM tbl_students WHERE name = 'Test Student Delta');

INSERT INTO tbl_teachers_students_m2m (teacher_id, student_id)
SELECT t.id, s.id
FROM tbl_teachers t
CROSS JOIN tbl_students s
WHERE t.deleted = 0
	AND t.email IN ('testgoogle@zion.com', 'testzoom@zion.com')
	AND s.name IN (
		'Test Student Alpha',
		'Test Student Beta',
		'Test Student Gamma',
		'Test Student Delta'
	)
	AND NOT EXISTS (
		SELECT 1 FROM tbl_teachers_students_m2m m
		WHERE m.teacher_id = t.id AND m.student_id = s.id
	);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_teachers_students_m2m
WHERE student_id IN (
	SELECT id FROM tbl_students
	WHERE name IN (
		'Test Student Alpha',
		'Test Student Beta',
		'Test Student Gamma',
		'Test Student Delta'
	)
);

DELETE FROM tbl_students
WHERE name IN (
	'Test Student Alpha',
	'Test Student Beta',
	'Test Student Gamma',
	'Test Student Delta'
);
-- +goose StatementEnd
