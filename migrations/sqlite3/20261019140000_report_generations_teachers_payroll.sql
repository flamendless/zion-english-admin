-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_report_generations_payroll (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	kind TEXT NOT NULL DEFAULT 'teacher' CHECK(kind IN ('teacher', 'summary', 'teachers_payroll')),
	teacher_id INTEGER REFERENCES tbl_teachers(id),
	start_date TEXT NOT NULL,
	end_date TEXT NOT NULL,
	content_hash TEXT NOT NULL,
	output_path TEXT NOT NULL,
	record_count INTEGER NOT NULL,
	generated_at TEXT NOT NULL DEFAULT (datetime('now')),
	CHECK (
		(kind = 'teacher' AND teacher_id IS NOT NULL)
		OR (kind IN ('summary', 'teachers_payroll') AND teacher_id IS NULL)
	)
);

INSERT INTO tbl_report_generations_payroll (
	id, kind, teacher_id, start_date, end_date, content_hash, output_path, record_count, generated_at
)
SELECT
	id, kind, teacher_id, start_date, end_date, content_hash, output_path, record_count, generated_at
FROM tbl_report_generations;

DROP TABLE tbl_report_generations;
ALTER TABLE tbl_report_generations_payroll RENAME TO tbl_report_generations;

CREATE UNIQUE INDEX idx_report_generations_teacher_period
	ON tbl_report_generations(teacher_id, start_date, end_date)
	WHERE kind = 'teacher';

CREATE UNIQUE INDEX idx_report_generations_summary_period
	ON tbl_report_generations(start_date, end_date)
	WHERE kind = 'summary';

CREATE UNIQUE INDEX idx_report_generations_teachers_payroll_period
	ON tbl_report_generations(start_date, end_date)
	WHERE kind = 'teachers_payroll';

PRAGMA foreign_keys=ON;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
PRAGMA foreign_keys=OFF;

CREATE TABLE tbl_report_generations_legacy (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	kind TEXT NOT NULL DEFAULT 'teacher' CHECK(kind IN ('teacher', 'summary')),
	teacher_id INTEGER REFERENCES tbl_teachers(id),
	start_date TEXT NOT NULL,
	end_date TEXT NOT NULL,
	content_hash TEXT NOT NULL,
	output_path TEXT NOT NULL,
	record_count INTEGER NOT NULL,
	generated_at TEXT NOT NULL DEFAULT (datetime('now')),
	CHECK (
		(kind = 'teacher' AND teacher_id IS NOT NULL)
		OR (kind = 'summary' AND teacher_id IS NULL)
	)
);

INSERT INTO tbl_report_generations_legacy (
	id, kind, teacher_id, start_date, end_date, content_hash, output_path, record_count, generated_at
)
SELECT
	id, kind, teacher_id, start_date, end_date, content_hash, output_path, record_count, generated_at
FROM tbl_report_generations
WHERE kind != 'teachers_payroll';

DROP TABLE tbl_report_generations;
ALTER TABLE tbl_report_generations_legacy RENAME TO tbl_report_generations;

CREATE UNIQUE INDEX idx_report_generations_teacher_period
	ON tbl_report_generations(teacher_id, start_date, end_date)
	WHERE kind = 'teacher';

CREATE UNIQUE INDEX idx_report_generations_summary_period
	ON tbl_report_generations(start_date, end_date)
	WHERE kind = 'summary';

PRAGMA foreign_keys=ON;
-- +goose StatementEnd
