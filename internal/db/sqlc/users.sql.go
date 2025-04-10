// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: users.sql

package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const clearActiveProject = `-- name: ClearActiveProject :exec
UPDATE users
SET
    active_project_id = NULL,
    updated_at = NOW()
WHERE id = $1
`

func (q *Queries) ClearActiveProject(ctx context.Context, id int32) error {
	_, err := q.db.Exec(ctx, clearActiveProject, id)
	return err
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (
    email,
    password_hash,
    name
) VALUES (
    $1, $2, $3
) RETURNING id, email, password_hash, name, created_at, updated_at, active_project_id
`

type CreateUserParams struct {
	Email        string      `json:"email"`
	PasswordHash string      `json:"password_hash"`
	Name         pgtype.Text `json:"name"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRow(ctx, createUser, arg.Email, arg.PasswordHash, arg.Name)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ActiveProjectID,
	)
	return i, err
}

const getActiveProject = `-- name: GetActiveProject :one
SELECT p.id, p.user_id, p.name, p.description, p.deadline, p.created_at, p.updated_at FROM projects p
JOIN users u ON p.id = u.active_project_id
WHERE u.id = $1
LIMIT 1
`

func (q *Queries) GetActiveProject(ctx context.Context, id int32) (Project, error) {
	row := q.db.QueryRow(ctx, getActiveProject, id)
	var i Project
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.Description,
		&i.Deadline,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getUser = `-- name: GetUser :one
SELECT id, email, password_hash, name, created_at, updated_at, active_project_id
FROM users
WHERE email = $1
LIMIT 1
`

func (q *Queries) GetUser(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRow(ctx, getUser, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ActiveProjectID,
	)
	return i, err
}

const setActiveProject = `-- name: SetActiveProject :exec
UPDATE users
SET
    active_project_id = $2,
    updated_at = NOW()
WHERE id = $1
`

type SetActiveProjectParams struct {
	ID              int32       `json:"id"`
	ActiveProjectID pgtype.Int4 `json:"active_project_id"`
}

func (q *Queries) SetActiveProject(ctx context.Context, arg SetActiveProjectParams) error {
	_, err := q.db.Exec(ctx, setActiveProject, arg.ID, arg.ActiveProjectID)
	return err
}

const updateUserEmail = `-- name: UpdateUserEmail :one
UPDATE users
SET 
    email = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, email, password_hash, name, created_at, updated_at, active_project_id
`

type UpdateUserEmailParams struct {
	ID    int32  `json:"id"`
	Email string `json:"email"`
}

func (q *Queries) UpdateUserEmail(ctx context.Context, arg UpdateUserEmailParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserEmail, arg.ID, arg.Email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ActiveProjectID,
	)
	return i, err
}

const updateUserPassword = `-- name: UpdateUserPassword :one
UPDATE users
SET 
    password_hash = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING id, email, password_hash, name, created_at, updated_at, active_project_id
`

type UpdateUserPasswordParams struct {
	ID           int32  `json:"id"`
	PasswordHash string `json:"password_hash"`
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) (User, error) {
	row := q.db.QueryRow(ctx, updateUserPassword, arg.ID, arg.PasswordHash)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.PasswordHash,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.ActiveProjectID,
	)
	return i, err
}
