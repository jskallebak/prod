-- +goose Up
-- SQL in this section is executed when the migration is applied

-- Add required columns to pomodoro_sessions table
ALTER TABLE pomodoro_sessions ADD COLUMN IF NOT EXISTS work_duration INTEGER NOT NULL DEFAULT 25;
ALTER TABLE pomodoro_sessions ADD COLUMN IF NOT EXISTS break_duration INTEGER NOT NULL DEFAULT 5;
ALTER TABLE pomodoro_sessions ADD COLUMN IF NOT EXISTS pause_time TIMESTAMPTZ NULL;
ALTER TABLE pomodoro_sessions ADD COLUMN IF NOT EXISTS total_pause_duration INTEGER NULL;
ALTER TABLE pomodoro_sessions ADD COLUMN IF NOT EXISTS actual_work_duration INTEGER NULL;
ALTER TABLE pomodoro_sessions ADD COLUMN IF NOT EXISTS note TEXT NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back

-- Remove added columns from pomodoro_sessions table
ALTER TABLE pomodoro_sessions DROP COLUMN IF EXISTS work_duration;
ALTER TABLE pomodoro_sessions DROP COLUMN IF EXISTS break_duration;
ALTER TABLE pomodoro_sessions DROP COLUMN IF EXISTS pause_time;
ALTER TABLE pomodoro_sessions DROP COLUMN IF EXISTS total_pause_duration;
ALTER TABLE pomodoro_sessions DROP COLUMN IF EXISTS actual_work_duration;
ALTER TABLE pomodoro_sessions DROP COLUMN IF EXISTS note; 