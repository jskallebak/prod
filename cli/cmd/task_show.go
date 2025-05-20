package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

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
		input, err := util.Input2Int(args[0])
		taskID, err := services.GetID(services.GetTaskMap, input)
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
		projectService := services.NewProjectService(queries)

		// Currently using a hardcoded user ID (1)
		// In a real app, you would get this from authentication
		userID := int32(1)

		// Get task details
		task, err := taskService.GetTask(context.Background(), int32(taskID), userID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to find task with ID %d: %v\n", taskID, err)
			return
		}

		// Get project name if task has a project
		var projectName string
		if task.ProjectID.Valid {
			project, err := projectService.GetProject(context.Background(), task.ProjectID.Int32, userID)
			if err == nil && project != nil {
				projectName = project.Name
			} else {
				projectName = fmt.Sprintf("ID %d", task.ProjectID.Int32)
			}
		}

		// Show task status with checkbox
		status := "[ ]"
		if task.CompletedAt.Valid {
			status = "[✓]"
		}
		fmt.Printf("%s #%s %s\n", status, input, task.Description)

		// Show task details with emojis and consistent formatting
		if task.CompletedAt.Valid {
			fmt.Printf("    ✅\tCompleted: %s\n", task.CompletedAt.Time.Format("2006-01-02 15:04"))
		}
		if task.Priority.Valid {
			// Convert priority letter to full name
			priorityName := "Unknown"
			switch task.Priority.String {
			case "H":
				priorityName = "High"
			case "M":
				priorityName = "Medium"
			case "L":
				priorityName = "Low"
			}
			fmt.Printf("    🔄\tPriority: %s\n", priorityName)
		} else {
			fmt.Printf("    🔄\tPriority: --\n")
		}
		if task.ProjectID.Valid {
			fmt.Printf("    📁\tProject: %s\n", projectName)
		} else {
			fmt.Printf("    📁\tProject: --\n")
		}
		if task.DueDate.Valid {
			fmt.Printf("    📅\tDue: %s\n", task.DueDate.Time.Format("Mon, Jan 2, 2006"))
		} else {
			fmt.Printf("    📅\tDue: --\n")
		}
		if len(task.Tags) > 0 {
			fmt.Printf("    🏷️\tTags: %s\n", strings.Join(task.Tags, ", "))
		} else {
			fmt.Printf("    🏷️\tTags: --\n")
		}
		fmt.Println()

		if task.StartDate.Valid {
			fmt.Printf("Start date: %s\n", task.StartDate.Time.Format("2006-01-02"))
		}

		if task.Recurrence.Valid {
			fmt.Printf("Recurrence: %s\n", task.Recurrence.String)
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
