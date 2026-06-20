-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_password_reset_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    teacher_id INTEGER REFERENCES tbl_teachers(id),
    reset_token TEXT,
    status TEXT NOT NULL CHECK(status IN (
        'requested', 'blocked', 'token_issued', 'completed', 'failed'
    )),
    event TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_password_reset_ip_created
    ON tbl_password_reset_events(ip_address, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_password_reset_token
    ON tbl_password_reset_events(reset_token);

-- +goose Down
DROP INDEX IF EXISTS idx_password_reset_token;
DROP INDEX IF EXISTS idx_password_reset_ip_created;
DROP TABLE IF EXISTS tbl_password_reset_events;
