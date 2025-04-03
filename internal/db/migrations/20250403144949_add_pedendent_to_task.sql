-- +goose Up
-- SQL in this section is executed when the migration is applied
ALTER TABLE tasks
ADD COLUMN dependent_task INTEGER REFERENCES tasks(id) ON DELETE SET NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
ALTER TABLE tasks
DROP COLUMN dependent_task;
