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

func (a AuthService) GetCurrentUser(ctx context.Context) (*sqlc.User, error) {
	token, err := auth.ReadToken()
	if err != nil {
		return nil, fmt.Errorf("failed to read the token: %w", err)
	}

	claim, err := auth.VerifyJWT(token)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	email := claim.Email
	user, err := a.queries.GetUser(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user in token not found in database: %w", err)
	}

	return &user, nil
}

func (a *AuthService) UpdateEmail(ctx context.Context, userID int32, newEmail string) (*sqlc.User, error) {
	params := sqlc.UpdateUserEmailParams{
		ID:    userID,
		Email: newEmail,
	}

	updatedUser, err := a.queries.UpdateUserEmail(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update email in database: %w", err)
	}

	token, err := auth.GenerateJWT(updatedUser.ID, updatedUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token: %w", err)
	}

	err = auth.StoreToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to store new token: %w", err)
	}

	return &updatedUser, nil
}

func (a *AuthService) UpdatePassword(ctx context.Context, userID int32, newPassword string) (*sqlc.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), COST)
	if err != nil {
		return nil, fmt.Errorf("error generate hash from password %w", err)
	}

	params := sqlc.UpdateUserPasswordParams{
		ID:           userID,
		PasswordHash: string(hash),
	}

	updatedUser, err := a.queries.UpdateUserPassword(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error updating the password in the DB: %w", err)
	}

	return &updatedUser, nil
}

func (a AuthService) GetHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), COST)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
