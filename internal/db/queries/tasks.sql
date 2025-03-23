-- name: CreateTask :one
INSERT INTO tasks (
    user_id,
    description,
    status,
    priority,
    due_date,
    start_date,
    project_id,
    recurrence,
    tags,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE email = $1
LIMIT 1;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: ListTasks :many
SELECT 
    id, 
    user_id,
    description, 
    status, 
    priority, 
    due_date, 
    start_date, 
    completed_at, 
    project_id,
    recurrence,
    tags,
    notes,
    created_at,
    updated_at
FROM 
    tasks
WHERE user_id = $1
AND (
    sqlc.narg(priority)::text IS NULL 
    OR priority = sqlc.narg(priority)
)
AND (
    sqlc.narg(project)::text IS NULL 
    OR project_id = sqlc.narg(project)::integer
);

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









-- name: CountTasks :one
SELECT 
    COUNT(*) AS total_tasks,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) AS pending_tasks,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) AS completed_tasks
FROM 
    tasks
WHERE 
    tasks.user_id = sqlc.arg(user_id)
    AND (
        sqlc.arg(status)::text IS NULL OR status = sqlc.arg(status)
    )
    AND (
        sqlc.arg(tags)::text[] IS NULL OR tags && sqlc.arg(tags)
    )
    AND (
        sqlc.arg(due_date)::timestamp WITH TIME ZONE IS NULL OR due_date <= sqlc.arg(due_date)
    )
    AND (
        sqlc.arg(project_name)::text IS NULL OR project_id IN (
            SELECT id FROM projects 
            WHERE name = sqlc.arg(project_name) AND user_id = sqlc.arg(user_id)
        )
    );

-- name: UpdateTask :one
UPDATE tasks
SET
    description = COALESCE($3, description),
    status = COALESCE($4, status),
    priority = COALESCE($5, priority),
    due_date = COALESCE($6, due_date),
    start_date = COALESCE($7, start_date),
    project_id = COALESCE($8, project_id),
    recurrence = COALESCE($9, recurrence),
    tags = COALESCE($10, tags),
    notes = COALESCE($11, notes),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: CompleteTask :one
UPDATE tasks
SET
    status = 'completed',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1 AND user_id = $2;


-- name: AddTaskDependency :exec
INSERT INTO task_dependencies (
    task_id,
    depends_on_id
) VALUES (
    $1, $2
);

-- name: RemoveTaskDependency :exec
DELETE FROM task_dependencies
WHERE task_id = $1 AND depends_on_id = $2;

-- name: GetTaskDependencies :many
SELECT t.* FROM tasks t
JOIN task_dependencies td ON t.id = td.depends_on_id
WHERE td.task_id = $1 AND t.user_id = $2
ORDER BY t.created_at DESC;

-- name: GetDependentTasks :many
SELECT t.* FROM tasks t
JOIN task_dependencies td ON t.id = td.task_id
WHERE td.depends_on_id = $1 AND t.user_id = $2
ORDER BY t.created_at DESC;

-- name: GetTasksWithinDateRange :many
SELECT * FROM tasks
WHERE user_id = $1
AND (
    (start_date IS NOT NULL AND start_date >= $2 AND start_date <= $3)
    OR (due_date IS NOT NULL AND due_date >= $2 AND due_date <= $3)
)
ORDER BY 
    COALESCE(start_date, due_date) ASC;

-- name: GetTasksByTag :many
SELECT * FROM tasks
WHERE user_id = $1
AND $2 = ANY(tags)
ORDER BY created_at DESC;

-- name: GetRecentlyCompletedTasks :many
SELECT * FROM tasks
WHERE user_id = $1
AND status = 'completed'
ORDER BY completed_at DESC
LIMIT $2;

-- name: UpdateTaskStatus :one
UPDATE tasks
SET
    status = $3,
    updated_at = NOW(),
    completed_at = CASE 
        WHEN $3 = 'completed' THEN NOW() 
        ELSE completed_at 
    END
WHERE id = $1 AND user_id = $2
RETURNING *;

