-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_teacher_leave_requests (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	teacher_id INTEGER NOT NULL,
	leave_type TEXT NOT NULL CHECK (leave_type IN ('sick', 'vacation', 'maternity', 'personal', 'bereavement', 'unpaid')),
	start_date TEXT NOT NULL,
	end_date TEXT NOT NULL,
	reason TEXT NOT NULL,
	status TEXT NOT NULL CHECK (status IN ('submitted', 'approved', 'rejected')),
	reviewed_by INTEGER,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	reviewed_at TEXT,
	FOREIGN KEY (teacher_id) REFERENCES tbl_teachers(id) ON DELETE CASCADE,
	FOREIGN KEY (reviewed_by) REFERENCES tbl_teachers(id)
);

CREATE INDEX IF NOT EXISTS idx_teacher_leave_requests_status
	ON tbl_teacher_leave_requests(status);

CREATE INDEX IF NOT EXISTS idx_teacher_leave_requests_teacher_id
	ON tbl_teacher_leave_requests(teacher_id);

-- +goose Down
DROP INDEX IF EXISTS idx_teacher_leave_requests_teacher_id;
DROP INDEX IF EXISTS idx_teacher_leave_requests_status;
DROP TABLE IF EXISTS tbl_teacher_leave_requests;