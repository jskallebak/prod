-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Add status column to pomodoro_sessions table if it doesn't exist
ALTER TABLE pomodoro_sessions ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'active';

-- Create pomodoro_config table if it doesn't exist
CREATE TABLE IF NOT EXISTS pomodoro_config (
    user_id INTEGER PRIMARY KEY REFERENCES users(id),
    work_duration INTEGER NOT NULL DEFAULT 25,
    break_duration INTEGER NOT NULL DEFAULT 5,
    long_break_duration INTEGER NOT NULL DEFAULT 15,
    long_break_interval INTEGER NOT NULL DEFAULT 4,
    auto_start_breaks BOOLEAN NOT NULL DEFAULT false,
    auto_start_pomodoros BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back

-- Remove status column from pomodoro_sessions table
ALTER TABLE pomodoro_sessions DROP COLUMN IF EXISTS status;

-- Drop pomodoro_config table
DROP TABLE IF EXISTS pomodoro_config;
