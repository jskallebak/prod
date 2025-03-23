package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
)

// TaskService handles business logic for tasks
type TaskService struct {
	queries *sqlc.Queries
}

// NewTaskService creates a new TaskService
func NewTaskService(queries *sqlc.Queries) *TaskService {
	return &TaskService{
		queries: queries,
	}
}

// TaskParams contains all possible parameters for creating a task
type TaskParams struct {
	Description string
	Priority    *string
	DueDate     *time.Time
	StartDate   *time.Time
	ProjectID   *int64
	Tags        []string
	Notes       *string
	Recurrence  *string
}

// CreateTask creates a new task with minimal required parameters
func (s *TaskService) CreateTask(ctx context.Context, userID int64, params TaskParams) (*sqlc.Task, error) {
	// Input validation - only description is required
	if params.Description == "" {
		return nil, fmt.Errorf("task description cannot be empty")
	}

	// Default status for new tasks
	status := "pending"

	// Convert Go types to pgtype types
	createParams := sqlc.CreateTaskParams{
		UserID: pgtype.Int4{
			Int32: int32(userID),
			Valid: true,
		},
		Description: params.Description,
		Status:      status,
		Tags:        params.Tags,
	}

	// Only set optional parameters if provided
	if params.Priority != nil {
		createParams.Priority = pgtype.Text{
			String: *params.Priority,
			Valid:  true,
		}
	}

	if params.DueDate != nil {
		createParams.DueDate = pgtype.Timestamptz{
			Time:  *params.DueDate,
			Valid: true,
		}
	}

	if params.StartDate != nil {
		createParams.StartDate = pgtype.Timestamptz{
			Time:  *params.StartDate,
			Valid: true,
		}
	}

	if params.ProjectID != nil {
		createParams.ProjectID = pgtype.Int4{
			Int32: int32(*params.ProjectID),
			Valid: true,
		}
	}

	if params.Notes != nil {
		createParams.Notes = pgtype.Text{
			String: *params.Notes,
			Valid:  true,
		}
	}

	if params.Recurrence != nil {
		createParams.Recurrence = pgtype.Text{
			String: *params.Recurrence,
			Valid:  true,
		}
	}

	// Call data layer
	task, err := s.queries.CreateTask(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return &task, nil
}

// GetTask retrieves a task by ID
func (s *TaskService) GetTask(ctx context.Context, taskID int32, userID int64) (*sqlc.Task, error) {
	task, err := s.queries.GetTask(ctx, sqlc.GetTaskParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: int32(userID),
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

// CompleteTask marks a task as completed
func (s *TaskService) CompleteTask(ctx context.Context, taskID int32, userID int64) (*sqlc.Task, error) {
	task, err := s.queries.CompleteTask(ctx, sqlc.CompleteTaskParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: int32(userID),
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to complete task: %w", err)
	}

	return &task, nil
}

// ListTasks retrieves tasks with optional filtering
func (s *TaskService) ListTasks(ctx context.Context, userID int, priority *string) ([]sqlc.Task, error) {
	// Create params object with userID being mandatory
	params := sqlc.ListTasksParams{
		UserID: pgtype.Int4{
			Int32: int32(userID),
			Valid: true,
		},
	}

	// Add priority if provided
	if priority != nil {
		params.Priority = pgtype.Text{
			String: *priority,
			Valid:  true,
		}
	}

	tasks, err := s.queries.ListTasks(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	return tasks, nil
}

// DeleteTask removes a task
func (s *TaskService) DeleteTask(ctx context.Context, taskID int32, userID int64) error {
	err := s.queries.DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: int32(userID),
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}
