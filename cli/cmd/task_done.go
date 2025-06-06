package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

// doneCmd represents the done command
var doneCmd = &cobra.Command{
	Use:   "done [task_id]",
	Short: "Mark a task as completed",
	Long: `Mark a task as completed in your productivity system.
	
For example:
  prod task done 5  # Marks task with ID 5 as completed`,
	// Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		inputs, err := util.ParseArgs(args)

		// Initialize DB connection
		dbpool, err := util.InitDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer dbpool.Close()

		// Create queries and services
		queries := sqlc.New(dbpool)
		taskService := services.NewTaskService(queries)
		authService := services.NewAuthService(queries)

		user, err := authService.GetCurrentUser(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "doneCmd: getCurrentUser: %v\n", err)
			return
		}

		for _, input := range inputs {
			taskID, err := services.GetID(services.GetTaskMap, input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "doneCmd: getID: %v\n", err)
				return
			}

			adaptedConfirm := func(ctx context.Context, taskID int32, userID int32, action string, ts *services.TaskService) error {
				return ConfirmCmd(ctx, taskID, userID, ActionType(action), ts)
			}

			err = services.RecursiveSubtasks(ctx, user.ID, taskID, taskService, "finish", input, adaptedConfirm, taskService.CompleteTask)
			if err != nil {
				fmt.Fprintf(os.Stderr, "doneCmd: %v\n", err)
				return
			}

			err = ConfirmCmd(ctx, taskID, user.ID, COMPLETE, taskService)
			if err != nil {
				fmt.Fprintf(os.Stderr, "doneCmd: ConfirmCmd: %v\n", err)
				return
			}

			// Complete the task
			completedTask, newTask, err := taskService.CompleteRecurringTask(ctx, taskID, user.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "doneCmd: taskService.CompleteRecurringTask: %v\n", err)
				return
			}

			fmt.Printf("Task %d marked as completed\n", input)
			fmt.Printf("Description: %s\n", completedTask.Description)
			fmt.Printf("Completed at: %s\n", completedTask.CompletedAt.Time.Format("2006-01-02 15:04:05"))

			// If this was a recurring task and a new task was created
			if newTask != nil {
				// Get a new task display ID for the newly created task
				taskMap, index, err := services.AppendToMap(newTask.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error adding new recurring task to map: %v\n", err)
				} else {
					err = services.MakeTaskMapFile(taskMap)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error updating task map: %v\n", err)
					}

					fmt.Printf("\nNext occurrence created as task %d\n", index)
					if newTask.DueDate.Valid {
						fmt.Printf("Due: %s\n", newTask.DueDate.Time.Format("2006-01-02"))
					}
				}
			}
		}
	},
}

func init() {
	taskCmd.AddCommand(doneCmd)
}
