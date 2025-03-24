-- Create pomodoro_sessions table
CREATE TABLE IF NOT EXISTS pomodoro_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id INTEGER REFERENCES tasks(id) ON DELETE SET NULL,
    status TEXT NOT NULL CHECK (status IN ('active', 'paused', 'completed', 'cancelled')),
    work_duration INTEGER NOT NULL, -- in minutes
    break_duration INTEGER NOT NULL, -- in minutes
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    pause_time TIMESTAMPTZ,
    total_pause_duration INTEGER, -- in seconds
    actual_work_duration INTEGER, -- in seconds
    note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS pomodoro_sessions_user_id_idx ON pomodoro_sessions(user_id);
CREATE INDEX IF NOT EXISTS pomodoro_sessions_task_id_idx ON pomodoro_sessions(task_id);
CREATE INDEX IF NOT EXISTS pomodoro_sessions_status_idx ON pomodoro_sessions(status);
CREATE INDEX IF NOT EXISTS pomodoro_sessions_start_time_idx ON pomodoro_sessions(start_time);

-- Create pomodoro_config table
CREATE TABLE IF NOT EXISTS pomodoro_config (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    work_duration INTEGER NOT NULL DEFAULT 25, -- in minutes
    break_duration INTEGER NOT NULL DEFAULT 5, -- in minutes
    long_break_duration INTEGER NOT NULL DEFAULT 15, -- in minutes
    long_break_interval INTEGER NOT NULL DEFAULT 4, -- after how many pomodoros
    auto_start_breaks BOOLEAN NOT NULL DEFAULT FALSE,
    auto_start_pomodoros BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
); 