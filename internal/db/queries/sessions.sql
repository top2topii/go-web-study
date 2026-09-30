-- name: CreateSession :exec
INSERT INTO sessions (token_hash, user_id, expires_at)
VALUES (?, ?, ?);

-- Returns the user owning a non-expired session.
-- name: GetSessionUser :one
SELECT sqlc.embed(users)
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.token_hash = sqlc.arg(token_hash)
  AND sessions.expires_at > sqlc.arg(now);

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token_hash = ?;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions
WHERE expires_at <= ?;
