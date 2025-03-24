package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
)

// ProjectService handles business logic for projects
type ProjectService struct {
	queries *sqlc.Queries
}

// NewProjectService creates a new ProjectService
func NewProjectService(queries *sqlc.Queries) *ProjectService {
	return &ProjectService{
		queries: queries,
	}
}

// ProjectParams contains all possible parameters for creating or updating a project
type ProjectParams struct {
	Name        string
	Description *string
	Deadline    *time.Time
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(ctx context.Context, userID int32, params ProjectParams) (*sqlc.Project, error) {
	// Input validation - name is required
	if params.Name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	// Convert Go types to pgtype types
	createParams := sqlc.CreateProjectParams{
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
		Name: params.Name,
	}

	// Only set optional parameters if provided
	if params.Description != nil {
		createParams.Description = pgtype.Text{
			String: *params.Description,
			Valid:  true,
		}
	}

	if params.Deadline != nil {
		createParams.Deadline = pgtype.Timestamptz{
			Time:  *params.Deadline,
			Valid: true,
		}
	}

	// Call data layer
	project, err := s.queries.CreateProject(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return &project, nil
}

// GetProject retrieves a project by ID
func (s *ProjectService) GetProject(ctx context.Context, projectID int32, userID int32) (*sqlc.Project, error) {
	project, err := s.queries.GetProject(ctx, sqlc.GetProjectParams{
		ID: projectID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// ListProjects retrieves all projects for a user
func (s *ProjectService) ListProjects(ctx context.Context, userID int32) ([]sqlc.Project, error) {
	projects, err := s.queries.ListProjects(ctx, pgtype.Int4{
		Int32: userID,
		Valid: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}

// UpdateProject updates a project
func (s *ProjectService) UpdateProject(ctx context.Context, projectID int32, userID int32, params ProjectParams) (*sqlc.Project, error) {
	updateParams := sqlc.UpdateProjectParams{
		ID: projectID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	}

	// Only update name if provided and not empty
	if params.Name != "" {
		updateParams.Name = params.Name
	}

	// Set optional fields if provided
	if params.Description != nil {
		updateParams.Description = pgtype.Text{
			String: *params.Description,
			Valid:  true,
		}
	}

	if params.Deadline != nil {
		updateParams.Deadline = pgtype.Timestamptz{
			Time:  *params.Deadline,
			Valid: true,
		}
	}

	project, err := s.queries.UpdateProject(ctx, updateParams)
	if err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return &project, nil
}

// DeleteProject deletes a project
func (s *ProjectService) DeleteProject(ctx context.Context, projectID int32, userID int32) error {
	err := s.queries.DeleteProject(ctx, sqlc.DeleteProjectParams{
		ID: projectID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// GetProjectTasks retrieves all tasks for a project
func (s *ProjectService) GetProjectTasks(ctx context.Context, projectID int32, userID int32) ([]sqlc.Task, error) {
	tasks, err := s.queries.GetProjectTasks(ctx, sqlc.GetProjectTasksParams{
		ProjectID: pgtype.Int4{
			Int32: projectID,
			Valid: true,
		},
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get project tasks: %w", err)
	}
	return tasks, nil
}

// RemoveTaskFromProject removes a task from a project
func (s *ProjectService) RemoveTaskFromProject(ctx context.Context, taskID int32, userID int32) (*sqlc.Task, error) {
	task, err := s.queries.RemoveTaskFromProject(ctx, sqlc.RemoveTaskFromProjectParams{
		ID: taskID,
		UserID: pgtype.Int4{
			Int32: userID,
			Valid: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove task from project: %w", err)
	}

	return &task, nil
}
