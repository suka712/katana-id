-- name: CreateUser :one
INSERT INTO users (username, email, email_verified)
VALUES ($1, $2, $3)
RETURNING *;

-- name: CreateOTP :exec
INSERT INTO otps (email, otp, expires_at)
VALUES ($1, $2, $3);

-- name: CreateSession :one
INSERT INTO sessions (email, expires_at)
VALUES ($1, $2)
RETURNING *;

-- name: CreateProvider :one
INSERT INTO providers (user_id, provider_name, provider_account_id)
VALUES ($1, $2, $3)
RETURNING *;