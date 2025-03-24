package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var removeTaskCmd = &cobra.Command{
	Use:   "remove [task-id]",
	Short: "Remove a task from its project",
	Long: `Remove a task from its associated project without deleting the task.

Examples:
  prod project task remove 5  # Remove task with ID 5 from its project`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		// Parse task ID
		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid task ID: %v\n", err)
			return
		}

		// Get authenticated user
		authService := services.NewAuthService(queries)
		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Println("You need to be logged in to remove tasks from projects")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Get task to verify it exists and has a project
		taskService := services.NewTaskService(queries)
		task, err := taskService.GetTask(context.Background(), int32(taskID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		// Check if task has a project
		if !task.ProjectID.Valid {
			fmt.Println("This task is not associated with any project")
			return
		}

		// Get project name for display
		projectService := services.NewProjectService(queries)
		project, err := projectService.GetProject(context.Background(), task.ProjectID.Int32, user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving project: %v\n", err)
			return
		}

		// Create updateParams with task's current values
		updateParams := sqlc.UpdateTaskParams{
			ID:          task.ID,
			UserID:      pgtype.Int4{Int32: user.ID, Valid: true},
			Description: task.Description,
			Status:      task.Status,
			Priority:    task.Priority,
			DueDate:     task.DueDate,
			StartDate:   task.StartDate,
			ProjectID:   pgtype.Int4{Valid: false}, // Set ProjectID to null
			Recurrence:  task.Recurrence,
			Tags:        task.Tags,
			Notes:       task.Notes,
		}

		// Update the task to remove project association
		updatedTask, err := queries.UpdateTask(context.Background(), updateParams)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error removing task from project: %v\n", err)
			return
		}

		fmt.Printf("Task '%s' (ID: %d) removed from project '%s' (ID: %d)\n",
			updatedTask.Description, updatedTask.ID, project.Name, project.ID)
	},
}

func init() {
	projectTaskCmd.AddCommand(removeTaskCmd)
}
