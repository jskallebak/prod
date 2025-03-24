-- name: GetUser :one
SELECT *
FROM users
WHERE email = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    name
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: UpdateUserEmail :one
UPDATE users
SET 
    email = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET 
    password_hash = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *; 