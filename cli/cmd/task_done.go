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
		input := args[0]
		taskID, err := getID(getTaskMap, input)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

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

		err = ConfirmCmd(context.Background(), input, taskID, user.ID, COMPLETE, taskService)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}

		// Complete the task
		completedTask, err := taskService.CompleteTask(context.Background(), taskID, user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marking task as completed: %v\n", err)
			return
		}

		fmt.Printf("Task %s marked as completed\n", input)
		fmt.Printf("Description: %s\n", completedTask.Description)
		fmt.Printf("Completed at: %s\n", completedTask.CompletedAt.Time.Format("2006-01-02 15:04:05"))
	},
}

func init() {
	taskCmd.AddCommand(doneCmd)
}
