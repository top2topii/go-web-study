-- name: CreateUser :one
INSERT INTO users (email, name, password_hash, role)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = ?;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id;
