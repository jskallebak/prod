package services

import (
	"github.com/jskallebak/prod/internal/db/sqlc"
	"golang.org/x/crypto/bcrypt"
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

func (a AuthService) login(email, password string) *sqlc.User {
	return nil
}

func (a AuthService) GetHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
