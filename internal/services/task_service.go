package services

import (
	"context"
	"fmt"
	"slices"
	"strings"
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
	ProjectID   *int32
	Tags        []string
	Notes       *string
	Recurrence  *string
	Dependent   int32
}

// CompleteRecurringTask completes a task and generates the next instance if it's recurring
func (s *TaskService) CompleteRecurringTask(ctx context.Context, taskID, userID int32) (*sqlc.Task, *sqlc.Task, error) {
	// Complete the current task
	completedTask, err := s.CompleteTask(ctx, taskID, userID)
	if err != nil {
		return nil, nil, err
	}

	// Check if it's a recurring task
	if completedTask.Recurrence.Valid && completedTask.Recurrence.String != "" {
		// Generate the next instance
		nextTask, err := GenerateNextTaskInstance(*completedTask, nil)
		if err != nil {
			// Log error but don't fail the completion
			return completedTask, nil, fmt.Errorf("failed to generate next recurring task: %w", err)
		}

		// Create the next task instance in the database
		createParams := sqlc.CreateTaskParams{
			UserID:      nextTask.UserID,
			Description: nextTask.Description,
			Status:      nextTask.Status,
			Priority:    nextTask.Priority,
			DueDate:     nextTask.DueDate,
			StartDate:   nextTask.StartDate,
			ProjectID:   nextTask.ProjectID,
			Recurrence:  nextTask.Recurrence,
			Tags:        nextTask.Tags,
			Notes:       nextTask.Notes,
			Dependent:   nextTask.Dependent,
		}

		createdTask, err := s.queries.CreateTask(ctx, createParams)
		if err != nil {
			return completedTask, nil, fmt.Errorf("failed to create next task instance: %w", err)
		}

		return completedTask, &createdTask, nil
	}

	// Not a recurring task, just return the completed task
	return completedTask, nil, nil
}

// CreateTask creates a new task with minimal required parameters
func (s *TaskService) CreateTask(ctx context.Context, userID int32, params TaskParams) (*sqlc.Task, error) {
	// Input validation - only description is required
	if params.Description == "" {
		return nil, fmt.Errorf("task description cannot be empty")
	}

	// Default status for new tasks
	status := "pending"

	// Convert Go types to pgtype types
	createParams := sqlc.CreateTaskParams{
		UserID: pgtype.Int4{
			Int32: userID,
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
			Int32: *params.ProjectID,
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

	if params.Dependent != 0 {
		createParams.Dependent = pgtype.Int4{
			Int32: params.Dependent,
			Valid: true,
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
func (s *TaskService) GetTask(ctx context.Context, taskID int32, userID int32) (*sqlc.Task, error) {
	task, err := s.queries.GetTask(ctx, sqlc.GetTaskParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

func (s *TaskService) PauseTask(ctx context.Context, taskID, userID int32) (*sqlc.Task, error) {
	task, err := s.queries.PauseTask(ctx, sqlc.PauseTaskParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update the task to active: %w", err)
	}
	return &task, nil
}

func (s *TaskService) StartTask(ctx context.Context, taskID, userID int32) (*sqlc.Task, error) {
	task, err := s.queries.StartTask(ctx, sqlc.StartTaskParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update task to active: %w", err)
	}
	return &task, nil
}

// CompleteTask marks a task as completed
func (s *TaskService) CompleteTask(ctx context.Context, taskID int32, userID int32) (*sqlc.Task, error) {
	task, err := s.queries.CompleteTask(ctx, sqlc.CompleteTaskParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update task to completed: %w", err)
	}

	return &task, nil
}

func (s *TaskService) ListTasks(ctx context.Context, userID int32, priority *string, project *string, tags []string, status []string, today bool) ([]sqlc.Task, error) {
	// Create params object with userID being mandatory
	params := sqlc.ListTasksParams{
		UserID: pgtype.Int4{
			Int32: userID,
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

	// Add project if provided
	if project != nil {
		params.Project = pgtype.Text{
			String: *project,
			Valid:  true,
		}
	}

	// Add status if provided
	if len(status) > 0 {
		params.Status = status
	}

	if len(tags) > 0 {
		params.Tags = tags
	}

	if today {
		params.TodayFilter = pgtype.Bool{
			Bool:  true,
			Valid: true,
		}
	}

	tasks, err := s.queries.ListTasks(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed ListTasks query: %w", err)
	}

	return tasks, nil
}

// DeleteTask removes a task
func (s *TaskService) DeleteTask(ctx context.Context, taskID int32, userID int32) (*sqlc.Task, error) {
	task, err := s.queries.DeleteTask(ctx, sqlc.DeleteTaskParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to delete task: %w", err)
	}

	return &task, nil
}

func (s TaskService) SetDue(ctx context.Context, userID int32, taskID int32, date *time.Time) (*sqlc.Task, error) {
	var params sqlc.SetTaskDueParams
	params.ID = taskID
	params.UserID = pgtype.Int4{
		Int32: userID,
		Valid: true,
	}

	if date != nil {
		params.Date = pgtype.Date{
			Time:  *date,
			Valid: true,
		}
	}

	task, err := s.queries.SetTaskDue(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("DueToday: failed to set due_date: %v", err)
	}

	return &task, nil

}

func (s *TaskService) AddTag(ctx context.Context, userID, taskID int32, tags []string) error {
	oldTags, err := s.GetTags(ctx, userID, taskID)
	if err != nil {
		return err
	}

	cleanedTags := []string{}
	for _, tag := range tags {
		cleaned := strings.TrimSpace(tag)
		if cleaned == "" {
			continue
		}
		cleanedTags = append(cleanedTags, tag)

	}

	tags = append(oldTags, cleanedTags...)
	params := sqlc.SetTagsParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
		Tags: tags,
	}
	err = s.queries.SetTags(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) GetTags(ctx context.Context, userID, taskID int32) ([]string, error) {
	params := sqlc.GetTagsParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	}

	tags, err := s.queries.GetTags(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("Failed to get tags: %w", err)
	}

	return tags, nil
}

func (s *TaskService) ClearTags(ctx context.Context, userID, taskID int32) error {
	params := sqlc.ClearTagsParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	}
	err := s.queries.ClearTags(ctx, params)
	if err != nil {
		return fmt.Errorf("Failed to clear tags: %w", err)
	}
	return nil
}

func (s *TaskService) RemoveTags(ctx context.Context, userID, taskID int32, tags []string) error {
	removeMap := make(map[string]struct{})
	for _, s := range tags {
		removeMap[s] = struct{}{}
	}

	oldTags, err := s.GetTags(ctx, userID, taskID)
	if err != nil {
		return err
	}

	result := slices.DeleteFunc(oldTags, func(s string) bool {
		_, found := removeMap[s]
		return found
	})

	params := sqlc.SetTagsParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
		Tags: result,
	}

	err = s.queries.SetTags(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) GetToday(ctx context.Context, userID int32) ([]sqlc.Task, error) {
	user := pgtype.Int4{
		Int32: userID,
		Valid: true,
	}
	tasks, err := s.queries.GetToday(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed GetToday query: %w", err)
	}
	return tasks, nil
}

func (s TaskService) GetDependent(ctx context.Context, userID int32, taskID int32) ([]sqlc.Task, error) {
	tasks, err := s.ListTasks(ctx, userID, nil, nil, nil, nil, false)
	if err != nil {
		return nil, err
	}

	task, err := s.GetTask(ctx, taskID, userID)
	if err != nil {
		return nil, fmt.Errorf("GetDependent: error getting the task: %w", err)
	}

	result := []sqlc.Task{}
	for _, t := range tasks {
		if t.Dependent.Int32 == task.ID {
			result = append(result, t)
		}
	}

	return result, nil
}
