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

-- name: SetActiveProject :exec
UPDATE users
SET
    active_project_id = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: GetActiveProject :one
SELECT p.* FROM projects p
JOIN users u ON p.id = u.active_project_id
WHERE u.id = $1
LIMIT 1;

-- name: ClearActiveProject :exec
UPDATE users
SET
    active_project_id = NULL,
    updated_at = NOW()
WHERE id = $1;
