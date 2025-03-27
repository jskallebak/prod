package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

const (
	PASSWORD_COST = 12
)

// UserService handles business logic for users
type UserService struct {
	queries *sqlc.Queries
}

// NewUserService creates a new UserService
func NewUserService(queries *sqlc.Queries) *UserService {
	return &UserService{
		queries: queries,
	}
}

// CreateUserParams contains the parameters for creating a new user
type CreateUserParams struct {
	Email    string
	Password string
	Name     *string
}

// CreateUser creates a new user with the provided parameters
func (s *UserService) CreateUser(ctx context.Context, params CreateUserParams) (*sqlc.User, error) {
	// Validate input
	if params.Email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	if params.Password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.Password), PASSWORD_COST)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Set up parameters for database query
	createParams := sqlc.CreateUserParams{
		Email:        params.Email,
		PasswordHash: string(hashedPassword),
	}

	// Add name if provided
	if params.Name != nil {
		createParams.Name = pgtype.Text{
			String: *params.Name,
			Valid:  true,
		}
	}

	// Create the user in the database
	user, err := s.queries.CreateUser(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (s *UserService) SetActiveProject(ctx context.Context, userID, projectID int32) error {
	params := sqlc.SetActiveProjectParams{
		ID: userID,
		ActiveProjectID: pgtype.Int4{
			Int32: projectID,
			Valid: true,
		},
	}

	err := s.queries.SetActiveProject(ctx, params)
	if err != nil {
		return err

	}
	return nil
}

func (s UserService) GetActiveProject(ctx context.Context, userID int32) (*sqlc.Project, error) {
	proj, err := s.queries.GetActiveProject(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &proj, nil
}

func (s *UserService) ClearActiveProject(ctx context.Context, userID int32) error {
	err := s.queries.ClearActiveProject(ctx, userID)
	if err != nil {
		return err
	}
	return nil
}
