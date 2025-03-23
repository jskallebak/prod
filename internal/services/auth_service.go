package services

import (
	"context"
	"fmt"

	"github.com/jskallebak/prod/internal/auth"
	"github.com/jskallebak/prod/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
)

const (
	COST = 12
)

type AuthService struct {
	queries *sqlc.Queries
}

// NewTaskService creates a new TaskService
func NewAuthService(queries *sqlc.Queries) *AuthService {
	return &AuthService{
		queries: queries,
	}
}

func (a AuthService) Login(ctx context.Context, email, password string) (*sqlc.User, error) {
	user, err := a.queries.GetUser(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid email or password: %w", err)
	}

	token, err := auth.GenerateJWT(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("could not generate authentication token: %w", err)
	}

	err = auth.StoreToken(token)
	if err != nil {
		return nil, fmt.Errorf("could not store authentication token: %w", err)
	}

	return &user, nil
}

func (a AuthService) GetHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), COST)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
