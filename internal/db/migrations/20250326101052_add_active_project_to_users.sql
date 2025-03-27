--l sk+goose Up
-- SQL in this section is executed when the migration is applied
ALTER TABLE users
ADD COLUMN active_project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
ALTER TABLE users
DROP COLUMN active_project_id;
 
