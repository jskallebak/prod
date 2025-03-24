-- name: CreateProject :one
INSERT INTO projects (
    user_id,
    name,
    description,
    deadline
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetProject :one
SELECT * FROM projects
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: ListProjects :many
SELECT * FROM projects
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET
    name = COALESCE($3, name),
    description = COALESCE($4, description),
    deadline = COALESCE($5, deadline),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteProject :exec
DELETE FROM projects
WHERE id = $1 AND user_id = $2;

-- name: GetProjectTasks :many
SELECT t.* FROM tasks t
WHERE t.project_id = $1 AND t.user_id = $2
ORDER BY t.created_at DESC;

-- name: RemoveTaskFromProject :one
UPDATE tasks
SET
    project_id = NULL,
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *; 