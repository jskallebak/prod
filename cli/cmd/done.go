package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

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
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse task ID from arguments
		taskID, err := strconv.Atoi(args[0])
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

		// Currently using a hardcoded user ID (1)
		// In a real app, you would get this from authentication
		userID := int32(1)

		// Complete the task
		completedTask, err := taskService.CompleteTask(context.Background(), int32(taskID), userID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marking task as completed: %v\n", err)
			return
		}

		fmt.Printf("Task %d marked as completed\n", completedTask.ID)
		fmt.Printf("Description: %s\n", completedTask.Description)
		fmt.Printf("Completed at: %s\n", completedTask.CompletedAt.Time.Format("2006-01-02 15:04:05"))
	},
}

func init() {
	taskCmd.AddCommand(doneCmd)
}
