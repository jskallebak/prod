package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DirPermissions  = os.FileMode(0700)
	FilePermissions = os.FileMode(0600)
	AppDirName      = ".prod"
	TokenFileName   = "auth.token"
)

func StoreToken(token string) error {
	if token == "" {
		return errors.New("cannot store empty token")
	}

	tokenPath, err := getTokenFilePath()
	if err != nil {
		return err
	}

	prodDir := filepath.Dir(tokenPath)
	err = os.MkdirAll(prodDir, DirPermissions)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	err = os.WriteFile(tokenPath, []byte(token), FilePermissions)
	if err != nil {
		return fmt.Errorf("failed to write the auth token: %w", err)
	}
	return nil
}

func ReadToken() (string, error) {
	tokenPath, err := getTokenFilePath()
	if err != nil {
		return "", err
	}

	token, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("not logged in (no token file found): %w", err)
		}
		return "", fmt.Errorf("error reading auth.token: %w", err)
	}
	return string(token), nil
}

func RemoveToken() error {
	tokenPath, err := getTokenFilePath()
	if err != nil {
		return err
	}

	err = os.Remove(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {

			return errors.New("Already logged out")
		}
		return fmt.Errorf("could not remove the token: %w", err)
	}

	return nil
}

func getTokenFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, AppDirName, TokenFileName), nil
}
