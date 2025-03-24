-- name: CreatePomodoroSession :one
INSERT INTO pomodoro_sessions (
    user_id,
    task_id,
    status,
    work_duration,
    break_duration,
    start_time,
    note,
    duration
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $4 /* Use work_duration for duration */
) RETURNING *;

-- name: GetActivePomodoroSession :one
SELECT * FROM pomodoro_sessions
WHERE user_id = $1 AND (status = 'active' OR status = 'paused')
ORDER BY created_at DESC
LIMIT 1;

-- name: StopPomodoroSession :one
UPDATE pomodoro_sessions
SET
    status = $3,
    end_time = $4,
    actual_work_duration = CASE
        WHEN total_pause_duration IS NOT NULL THEN
            EXTRACT(EPOCH FROM ($4 - start_time)) - total_pause_duration
        ELSE
            EXTRACT(EPOCH FROM ($4 - start_time))
    END
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: PausePomodoroSession :one
UPDATE pomodoro_sessions
SET
    status = $3,
    pause_time = $4
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: ResumePomodoroSession :one
UPDATE pomodoro_sessions
SET
    status = $3,
    pause_time = NULL,
    total_pause_duration = COALESCE(total_pause_duration, 0) + $4
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: AttachTaskToPomodoro :one
UPDATE pomodoro_sessions
SET
    task_id = $3
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DetachTaskFromPomodoro :one
UPDATE pomodoro_sessions
SET
    task_id = NULL
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetPomodoroSession :one
SELECT * FROM pomodoro_sessions
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: ListPomodoroSessions :many
SELECT * FROM pomodoro_sessions
WHERE user_id = $1
  AND (task_id = $2 OR $2 IS NULL)
  AND (start_time >= $3 OR $3 IS NULL)
  AND (start_time <= $4 OR $4 IS NULL)
ORDER BY start_time DESC
LIMIT $5;

-- name: GetPomodoroStats :one
SELECT
  COUNT(*) AS total_sessions,
  COUNT(CASE WHEN ps.status = 'completed' THEN 1 END) AS completed_sessions,
  COUNT(CASE WHEN ps.status = 'cancelled' THEN 1 END) AS cancelled_sessions,
  SUM(ps.work_duration) AS total_work_mins,
  SUM(ps.break_duration) AS total_break_mins,
  SUM(ps.work_duration) + SUM(ps.break_duration) AS total_duration_mins,
  AVG(ps.work_duration) AS avg_duration_mins,
  (
    SELECT DATE(sub_ps.start_time)
    FROM pomodoro_sessions sub_ps
    WHERE sub_ps.user_id = $1
      AND (sub_ps.task_id = $2 OR $2 IS NULL)
      AND (sub_ps.start_time >= $3 OR $3 IS NULL)
      AND (sub_ps.start_time <= $4 OR $4 IS NULL)
    GROUP BY DATE(sub_ps.start_time)
    ORDER BY COUNT(*) DESC
    LIMIT 1
  ) AS most_productive_day,
  (
    SELECT EXTRACT(HOUR FROM sub_ps.start_time)::int
    FROM pomodoro_sessions sub_ps
    WHERE sub_ps.user_id = $1
      AND (sub_ps.task_id = $2 OR $2 IS NULL)
      AND (sub_ps.start_time >= $3 OR $3 IS NULL)
      AND (sub_ps.start_time <= $4 OR $4 IS NULL)
    GROUP BY EXTRACT(HOUR FROM sub_ps.start_time)
    ORDER BY COUNT(*) DESC
    LIMIT 1
  ) AS most_productive_hour
FROM pomodoro_sessions ps
WHERE ps.user_id = $1
  AND (ps.task_id = $2 OR $2 IS NULL)
  AND (ps.start_time >= $3 OR $3 IS NULL)
  AND (ps.start_time <= $4 OR $4 IS NULL);

-- name: GetPomodoroConfig :one
SELECT * FROM pomodoro_config
WHERE user_id = $1
LIMIT 1;

-- name: UpsertPomodoroConfig :one
INSERT INTO pomodoro_config (
    user_id,
    work_duration,
    break_duration,
    long_break_duration,
    long_break_interval,
    auto_start_breaks,
    auto_start_pomodoros
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (user_id)
DO UPDATE SET
    work_duration = $2,
    break_duration = $3,
    long_break_duration = $4,
    long_break_interval = $5,
    auto_start_breaks = $6,
    auto_start_pomodoros = $7,
    updated_at = NOW()
RETURNING *; 