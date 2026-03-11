-- name: UpsertUserAgent :one
INSERT INTO tbl_useragents (
    user_agent,
    browser,
    browser_version,
    os,
    device,
    created_at,
    updated_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    datetime('now'),
    datetime('now')
)
ON CONFLICT(user_agent) DO UPDATE SET
    browser = excluded.browser,
    browser_version = excluded.browser_version,
    os = excluded.os,
    device = excluded.device,
    updated_at = datetime('now')
RETURNING id;

-- name: GetUserAgentByID :one
SELECT
    id,
    user_agent,
    browser,
    browser_version,
    os,
    device,
    created_at,
    updated_at
FROM tbl_useragents
WHERE id = ?
LIMIT 1;

-- name: GetUserAgentByUserAgent :one
SELECT
    id,
    user_agent,
    browser,
    browser_version,
    os,
    device,
    created_at,
    updated_at
FROM tbl_useragents
WHERE user_agent = ?
LIMIT 1;

-- name: CreateAccess :one
INSERT INTO tbl_accesses (
    teacher_id,
    login_at,
    useragent_id,
    created_at,
    updated_at
) VALUES (
    ?,
    datetime('now'),
    ?,
    datetime('now'),
    datetime('now')
) RETURNING id;

-- name: UpdateAccessLogout :one
UPDATE tbl_accesses
SET
    logout_at = datetime('now'),
    updated_at = datetime('now')
WHERE
    id = ?
RETURNING id;

-- name: GetAccessByID :one
SELECT
    id,
    teacher_id,
    login_at,
    logout_at,
    useragent_id,
    created_at,
    updated_at
FROM tbl_accesses
WHERE id = ?
LIMIT 1;
