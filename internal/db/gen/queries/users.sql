-- name: CreateUser :exec
INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3);

-- name: GetUserByUsername :one
SELECT username, password_hash FROM users WHERE username = $1;
