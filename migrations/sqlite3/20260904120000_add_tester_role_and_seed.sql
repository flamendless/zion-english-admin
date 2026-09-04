-- +goose Up
-- +goose StatementBegin
CREATE TABLE tbl_teacher_roles_new (
	teacher_id INTEGER NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('teacher', 'admin', 'developer', 'tester')),
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (teacher_id, role),
	FOREIGN KEY (teacher_id) REFERENCES tbl_teachers(id) ON DELETE CASCADE
);

INSERT INTO tbl_teacher_roles_new (teacher_id, role, created_at)
SELECT teacher_id, role, created_at FROM tbl_teacher_roles;

DROP INDEX IF EXISTS idx_teacher_roles_teacher_id;
DROP TABLE tbl_teacher_roles;
ALTER TABLE tbl_teacher_roles_new RENAME TO tbl_teacher_roles;
CREATE INDEX IF NOT EXISTS idx_teacher_roles_teacher_id ON tbl_teacher_roles(teacher_id);

INSERT INTO tbl_teachers (
	first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class,
	currency, drive_url, sex, password, status
)
SELECT
	'Test', '', 'Google', '1990-01-01', 'Zion English Test Account', '2026-01-01',
	'+639170000001', 'testgoogle@zion.com', NULL, '#B9D283', 0,
	'PHP', '', NULL, '$2a$10$Q6dY2WTBvMmN5owPIUGB6eklC0mH7AoeOjRaG9VfEZcoMNwtcpP4G', 'approved'
WHERE NOT EXISTS (
	SELECT 1 FROM tbl_teachers WHERE email = 'testgoogle@zion.com' AND deleted = 0
);

INSERT INTO tbl_teacher_roles (teacher_id, role)
SELECT id, 'tester' FROM tbl_teachers
WHERE email = 'testgoogle@zion.com' AND deleted = 0
	AND NOT EXISTS (
		SELECT 1 FROM tbl_teacher_roles tr
		WHERE tr.teacher_id = tbl_teachers.id AND tr.role = 'tester'
	);

INSERT INTO tbl_teachers (
	first_name, middle_name, last_name, birthdate, address, joining_date,
	mobile_number, email, certifications, assigned_color, rate_per_class,
	currency, drive_url, sex, password, status
)
SELECT
	'Test', '', 'Zoom', '1990-01-01', 'Zion English Test Account', '2026-01-01',
	'+639170000002', 'testzoom@zion.com', NULL, '#B9D283', 0,
	'PHP', '', NULL, '$2a$10$1E3PD9pMZU/eNmfAk2sQUON74B5A7EQoV7T.T12rYHCFgmQCQuFRW', 'approved'
WHERE NOT EXISTS (
	SELECT 1 FROM tbl_teachers WHERE email = 'testzoom@zion.com' AND deleted = 0
);

INSERT INTO tbl_teacher_roles (teacher_id, role)
SELECT id, 'tester' FROM tbl_teachers
WHERE email = 'testzoom@zion.com' AND deleted = 0
	AND NOT EXISTS (
		SELECT 1 FROM tbl_teacher_roles tr
		WHERE tr.teacher_id = tbl_teachers.id AND tr.role = 'tester'
	);

UPDATE tbl_feature_flags
SET visible_roles = visible_roles || ',tester'
WHERE key IN ('integration.zoom', 'integration.google_calendar')
	AND visible_roles NOT LIKE '%tester%';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM tbl_teacher_roles
WHERE teacher_id IN (
	SELECT id FROM tbl_teachers
	WHERE email IN ('testgoogle@zion.com', 'testzoom@zion.com')
);

DELETE FROM tbl_teachers
WHERE email IN ('testgoogle@zion.com', 'testzoom@zion.com');

UPDATE tbl_feature_flags
SET visible_roles = REPLACE(visible_roles, ',tester', '')
WHERE key IN ('integration.zoom', 'integration.google_calendar');

UPDATE tbl_feature_flags
SET visible_roles = REPLACE(visible_roles, 'tester,', '')
WHERE key IN ('integration.zoom', 'integration.google_calendar');

UPDATE tbl_feature_flags
SET visible_roles = REPLACE(visible_roles, 'tester', '')
WHERE key IN ('integration.zoom', 'integration.google_calendar');

CREATE TABLE tbl_teacher_roles_old (
	teacher_id INTEGER NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('teacher', 'admin', 'developer')),
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	PRIMARY KEY (teacher_id, role),
	FOREIGN KEY (teacher_id) REFERENCES tbl_teachers(id) ON DELETE CASCADE
);

INSERT INTO tbl_teacher_roles_old (teacher_id, role, created_at)
SELECT teacher_id, role, created_at FROM tbl_teacher_roles
WHERE role != 'tester';

DROP INDEX IF EXISTS idx_teacher_roles_teacher_id;
DROP TABLE tbl_teacher_roles;
ALTER TABLE tbl_teacher_roles_old RENAME TO tbl_teacher_roles;
CREATE INDEX IF NOT EXISTS idx_teacher_roles_teacher_id ON tbl_teacher_roles(teacher_id);
-- +goose StatementEnd
