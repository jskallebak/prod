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
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputID := args[0]
		taskMap, err := getTaskMap()
		if err != nil {
			taskMap = map[int]int32{}
		}
		taskID, err := getTaskID(taskMap, inputID)

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

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			return
		}

		// Complete the task
		completedTask, err := taskService.CompleteTask(context.Background(), int32(taskID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marking task as completed: %v\n", err)
			return
		}

		fmt.Printf("Task %sargs[0] marked as completed\n", inputID)
		fmt.Printf("Description: %s\n", completedTask.Description)
		fmt.Printf("Completed at: %s\n", completedTask.CompletedAt.Time.Format("2006-01-02 15:04:05"))
	},
}

func init() {
	taskCmd.AddCommand(doneCmd)
}
