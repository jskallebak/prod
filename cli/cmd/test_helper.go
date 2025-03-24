package cmd

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/stretchr/testify/mock"
)

// MockDB is a mock database for testing
type MockDB struct {
	mock.Mock
}

// Task represents a testing implementation of task methods
type TaskService interface {
	GetTask(ctx context.Context, id int32, userID int32) (*sqlc.Task, error)
	CompleteTask(ctx context.Context, id int32, userID int32) (*sqlc.Task, error)
	DeleteTask(ctx context.Context, id int32, userID int32) error
	UpdateTask(ctx context.Context, params sqlc.UpdateTaskParams) (*sqlc.Task, error)
}

// CreateTestTask creates a task for testing
func CreateTestTask(id int32, description string) sqlc.Task {
	return sqlc.Task{
		ID:          id,
		UserID:      pgtype.Int4{Int32: 1, Valid: true},
		Description: description,
		Status:      "TODO",
		Priority:    pgtype.Text{String: "M", Valid: true},
		Tags:        []string{"test"},
		CreatedAt:   pgtype.Timestamptz{Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Valid: true},
	}
}
