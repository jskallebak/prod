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

var confirmDelete bool

// deleteTaskCmd represents the delete command for tasks
var deleteTaskCmd = &cobra.Command{
	Use:   "delete [task_id]",
	Short: "Delete a task",
	Long: `Delete a task from your productivity system.
	
For example:
  prod task delete 5         # Prompts for confirmation
  prod task delete 5 --yes   # Deletes without confirmation`,
	// Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		// Parse task ID from arguments
		inputs, err := util.ParseArgs(args)

		// Initialize DB connection
		dbpool, err := util.InitDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer dbpool.Close()

		// Create queries and service
		queries := sqlc.New(dbpool)
		taskService := services.NewTaskService(queries)
		authService := services.NewAuthService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get user %v\n", err)
			return
		}

		adaptedConfirm := func(ctx context.Context, taskID int32, userID int32, action string, ts *services.TaskService) error {
			return ConfirmCmd(ctx, taskID, userID, ActionType(action), ts)
		}

		for _, input := range inputs {
			taskID, err := services.GetID(services.GetTaskMap, input)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}

			err = services.RecursiveSubtasks(ctx, user.ID, taskID, taskService, "delete", input, adaptedConfirm, taskService.DeleteTask)
			if err != nil {
				fmt.Fprintf(os.Stderr, "DeleteCmd: %v", err)
				return
			}

			if !confirmDelete {
				err = adaptedConfirm(ctx, taskID, user.ID, string(DELETE), taskService)
				if err != nil {
					fmt.Fprintf(os.Stderr, "%s\n", err)
					return
				}
			}

			_, err = taskService.DeleteTask(ctx, taskID, user.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "task_delete: Error deleting task: %v\n", err)
				return
			}

			err = services.RemoveFromMap(input)
			if err != nil {
				fmt.Println(err)
			}

			fmt.Printf("Task %d deleted successfully\n", input)
		}
	},
}

func init() {
	taskCmd.AddCommand(deleteTaskCmd)

	// Define flags for the delete command
	deleteTaskCmd.Flags().BoolVar(&confirmDelete, "yes", false, "Delete without confirmation")
}
