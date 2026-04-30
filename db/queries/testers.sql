-- name: CreateSignup :exec
INSERT INTO tester_signups (email, name, platform)
VALUES ($1, $2, $3)
    ON CONFLICT (email) DO NOTHING;

-- name: ListTesters :many
SELECT id, email, name, platform, status, created_at, approved_at, rejected_at, invited_at
FROM tester_signups
ORDER BY created_at DESC;

-- name: FindTesterByID :one
SELECT id, email, name, platform, status, created_at, approved_at, rejected_at, invited_at
FROM tester_signups
WHERE id = $1;

-- name: ApproveTester :exec
UPDATE tester_signups
SET status = 'approved',
    approved_at = NOW(),
    rejected_at = NULL
WHERE id = $1;

-- name: RejectTester :exec
UPDATE tester_signups
SET status = 'rejected',
    rejected_at = NOW()
WHERE id = $1;

-- name: MarkTesterInvited :exec
UPDATE tester_signups
SET status = 'invited',
    invited_at = NOW()
WHERE id = $1;