-- name: IsAdminWhitelisted :one
SELECT EXISTS (
    SELECT 1 FROM admin_whitelist WHERE email = $1
);

-- name: CreateMagicLink :exec
INSERT INTO magic_links (id, email, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: FindValidMagicLinkForUpdate :one
SELECT email
FROM magic_links
WHERE token_hash = $1
  AND used_at IS NULL
  AND expires_at > NOW()
    FOR UPDATE;

-- name: MarkMagicLinkUsed :exec
UPDATE magic_links
SET used_at = NOW()
WHERE token_hash = $1;

-- name: CreateAdminSession :exec
INSERT INTO admin_sessions (id, email, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: FindValidSession :one
SELECT email
FROM admin_sessions
WHERE token_hash = $1
  AND expires_at > NOW();

-- name: DeleteSession :exec
DELETE FROM admin_sessions
WHERE token_hash = $1;