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
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse task ID from arguments
		input := args[0]
		taskID, err := getID(getTaskMap, input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid task ID\n")
			return
		}

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
			fmt.Fprintf(os.Stderr, "Failed to get user %v", err)
		}

		// Get task info for confirmation
		task, err := taskService.GetTask(context.Background(), int32(taskID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to find task with ID %d: %v\n", taskID, err)
			return
		}

		// Confirm deletion unless --yes flag is used
		if !confirmDelete {
			fmt.Printf("You are about to delete task %s: \"%s\"\n", input, task.Description)
			fmt.Print("Are you sure? (y/N): ")
			var confirmation string
			fmt.Scanln(&confirmation)
			if confirmation != "y" && confirmation != "Y" {
				fmt.Println("Deletion cancelled")
				return
			}
		}

		// Delete the task
		err = taskService.DeleteTask(context.Background(), int32(taskID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting task: %v\n", err)
			return
		}

		fmt.Printf("Task %s deleted successfully\n", input)
	},
}

func init() {
	taskCmd.AddCommand(deleteTaskCmd)

	// Define flags for the delete command
	deleteTaskCmd.Flags().BoolVar(&confirmDelete, "yes", false, "Delete without confirmation")
}
