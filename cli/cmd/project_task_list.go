package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var listTasksCmd = &cobra.Command{
	Use:   "list [project-id]",
	Short: "List tasks in a project",
	Long: `List all tasks associated with a project.

Examples:
  prod project task list 1  # List all tasks in project with ID 1`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		// Parse project ID
		projectID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid project ID: %v\n", err)
			return
		}

		// Get authenticated user
		authService := services.NewAuthService(queries)
		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Println("You need to be logged in to list project tasks")
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

		// Get tasks for this project
		tasks, err := projectService.GetProjectTasks(context.Background(), int32(projectID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving project tasks: %v\n", err)
			return
		}

		// Display project info and task count
		fmt.Printf("Project: %s (ID: %d)\n", project.Name, project.ID)

		if len(tasks) == 0 {
			fmt.Println("No tasks found in this project")
			fmt.Printf("Use 'prod task create \"Task description\" --project %d' to add a task\n", projectID)
			fmt.Printf("Or 'prod project task add %d [task-id]' to add an existing task\n", projectID)
			return
		}

		fmt.Printf("Found %d tasks:\n\n", len(tasks))

		// Display tasks with status, priority, and due date
		for i, task := range tasks {
			fmt.Printf("%d. [%s] %s", i+1, task.Status, task.Description)

			if task.Priority.Valid {
				fmt.Printf(" (Priority: %s)", task.Priority.String)
			}

			if task.DueDate.Valid {
				fmt.Printf(" (Due: %s)", task.DueDate.Time.Format("2006-01-02"))
			}

			fmt.Println()
		}
	},
}

func init() {
	projectTaskCmd.AddCommand(listTasksCmd)
}
