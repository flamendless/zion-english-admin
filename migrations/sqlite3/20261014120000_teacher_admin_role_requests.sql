-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_teacher_admin_role_requests (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	target_teacher_id INTEGER NOT NULL,
	requested_by_teacher_id INTEGER NOT NULL,
	status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
	reviewed_by INTEGER,
	created_at TEXT NOT NULL DEFAULT (datetime('now')),
	reviewed_at TEXT,
	FOREIGN KEY (target_teacher_id) REFERENCES tbl_teachers(id) ON DELETE CASCADE,
	FOREIGN KEY (requested_by_teacher_id) REFERENCES tbl_teachers(id),
	FOREIGN KEY (reviewed_by) REFERENCES tbl_teachers(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_teacher_admin_role_requests_pending_target
	ON tbl_teacher_admin_role_requests(target_teacher_id)
	WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_teacher_admin_role_requests_status
	ON tbl_teacher_admin_role_requests(status);

-- +goose Down
DROP INDEX IF EXISTS idx_teacher_admin_role_requests_status;
DROP INDEX IF EXISTS idx_teacher_admin_role_requests_pending_target;
DROP TABLE IF EXISTS tbl_teacher_admin_role_requests;
