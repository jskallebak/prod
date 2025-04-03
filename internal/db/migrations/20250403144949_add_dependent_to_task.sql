-- +goose Up
-- SQL in this section is executed when the migration is applied
ALTER TABLE tasks
ADD COLUMN dependent INTEGER REFERENCES tasks(id);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
ALTER TABLE tasks
DROP COLUMN dependent;
