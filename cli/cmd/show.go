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

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show [task_id]",
	Short: "Show detailed task information",
	Long: `Show detailed information about a specific task.
	
For example:
  prod task show 5  # Shows details for task with ID 5`,
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

		// Get task details
		task, err := taskService.GetTask(context.Background(), int32(taskID), userID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to find task with ID %d: %v\n", taskID, err)
			return
		}

		// Display task details
		fmt.Printf("Task ID: %d\n", task.ID)
		fmt.Printf("Description: %s\n", task.Description)
		fmt.Printf("Status: %s\n", task.Status)

		if task.Priority.Valid {
			fmt.Printf("Priority: %s\n", task.Priority.String)
		} else {
			fmt.Println("Priority: Not set")
		}

		if task.DueDate.Valid {
			fmt.Printf("Due date: %s\n", task.DueDate.Time.Format("2006-01-02"))
		} else {
			fmt.Println("Due date: Not set")
		}

		if task.StartDate.Valid {
			fmt.Printf("Start date: %s\n", task.StartDate.Time.Format("2006-01-02"))
		}

		if task.CompletedAt.Valid {
			fmt.Printf("Completed at: %s\n", task.CompletedAt.Time.Format("2006-01-02 15:04:05"))
		}

		if task.ProjectID.Valid {
			fmt.Printf("Project ID: %d\n", task.ProjectID.Int32)
		}

		if task.Recurrence.Valid {
			fmt.Printf("Recurrence: %s\n", task.Recurrence.String)
		}

		if len(task.Tags) > 0 {
			fmt.Printf("Tags: %v\n", task.Tags)
		} else {
			fmt.Println("Tags: None")
		}

		if task.Notes.Valid && task.Notes.String != "" {
			fmt.Printf("Notes: %s\n", task.Notes.String)
		}

		fmt.Printf("Created at: %s\n", task.CreatedAt.Time.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated at: %s\n", task.UpdatedAt.Time.Format("2006-01-02 15:04:05"))
	},
}

func init() {
	taskCmd.AddCommand(showCmd)
}
