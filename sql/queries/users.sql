-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: GetAllUsers :many
SELECT * FROM users
ORDER BY created_at;

-- name: ResetUsersTable :exec
DELETE FROM users;
