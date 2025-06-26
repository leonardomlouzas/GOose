-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAllChirps :many
SELECT * FROM chirps
ORDER BY created_at;

-- name: GetOneChirp :one
SELECT * FROM chirps
WHERE id = $1;

-- name: ResetChirpsTable :exec
DELETE FROM chirps;