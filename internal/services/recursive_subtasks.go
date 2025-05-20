package services

import (
	"context"
	"fmt"

	"github.com/jskallebak/prod/internal/db/sqlc"
)

func RecursiveSubtasks(
	ctx context.Context,
	userID int32,
	taskID int32,
	ts *TaskService,
	action string,
	input int,
	confirmFunc func(context.Context, int32, int32, string, *TaskService) error,
	executeFunc func(context.Context, int32, int32) (*sqlc.Task, error),
) error {
	subtasks, err := ts.GetDependent(ctx, userID, taskID)
	if err != nil {
		return fmt.Errorf("RecursiveSubtasks: Error with GetDependent: %v", err)
	}

	if len(subtasks) != 0 {
		fmt.Printf("The task has subtask(s), confirm to %s\n", action)
		for _, st := range subtasks {
			err = confirmFunc(ctx, st.ID, userID, action, ts)
			if err != nil {
				return fmt.Errorf("RecursiveSubtasks: Error with ConfirmCmd: %v", err)
			}

			_, err = executeFunc(ctx, st.ID, userID)
			if err != nil {
				return fmt.Errorf("RecursiveSubtasks: Error with DeleteTask: %v", err)
			}
			fmt.Printf("Task %d deleted successfully\n", input)
		}
	}

	return nil
}
