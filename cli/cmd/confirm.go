package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/jskallebak/prod/internal/services"
)

func ConfirmCmd(ctx context.Context, taskID, userID int32, action ActionType, ts *services.TaskService) error {
	// Get task info for confirmation
	task, err := ts.GetTask(ctx, taskID, userID)
	if err != nil {
		return fmt.Errorf("error: Failed to find task with ID %d: %v", taskID, err)
	}

	// Confirm deletion unless --yes flag is used
	fmt.Printf("You are about to %s task %d: \"%s\"\n", action, taskID, task.Description)
	fmt.Print("Are you sure? (y/N): ")
	var confirmation string
	fmt.Scanln(&confirmation)
	if confirmation != "y" && confirmation != "Y" {
		return errors.New("complete cancelled")
	}
	return nil
}
