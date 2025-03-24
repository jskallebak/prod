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

var addTaskCmd = &cobra.Command{
	Use:   "add [project-id] [task-id]",
	Short: "Add a task to a project",
	Long: `Add an existing task to a project.

Examples:
  prod project task add 1 5  # Add task with ID 5 to project with ID 1`,

	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		// Parse project ID and task ID
		projectID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid project ID: %v\n", err)
			return
		}

		taskID, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid task ID: %v\n", err)
			return
		}

		// Get authenticated user
		authService := services.NewAuthService(queries)
		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Println("You need to be logged in to add tasks to projects")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Verify project exists and belongs to user
		projectService := services.NewProjectService(queries)
		project, err := projectService.GetProject(context.Background(), int32(projectID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		// Verify task exists and belongs to user
		taskService := services.NewTaskService(queries)
		task, err := taskService.GetTask(context.Background(), int32(taskID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		// Add task to project by updating task's project_id
		projectIDPgtype := pgtype.Int4{Int32: int32(projectID), Valid: true}

		// Create updateParams with minimal required fields
		updateParams := sqlc.UpdateTaskParams{
			ID:          task.ID,
			UserID:      pgtype.Int4{Int32: user.ID, Valid: true},
			Description: task.Description,
			Status:      task.Status,
			Priority:    task.Priority,
			DueDate:     task.DueDate,
			StartDate:   task.StartDate,
			ProjectID:   projectIDPgtype,
			Recurrence:  task.Recurrence,
			Tags:        task.Tags,
			Notes:       task.Notes,
		}

		// Call the UpdateTask method directly from queries
		updatedTask, err := queries.UpdateTask(context.Background(), updateParams)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding task to project: %v\n", err)
			return
		}

		fmt.Printf("Task '%s' (ID: %d) added to project '%s' (ID: %d)\n",
			updatedTask.Description, updatedTask.ID, project.Name, project.ID)
	},
}

func init() {
	projectTaskCmd.AddCommand(addTaskCmd)
}
