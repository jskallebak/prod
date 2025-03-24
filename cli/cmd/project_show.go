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

var showProjectCmd = &cobra.Command{
	Use:   "show [project-id]",
	Short: "Show project details and summary",
	Long: `Show detailed information about a project and a summary of its tasks.

Example:
  prod project show 1

This command shows project details and a summary of associated tasks.`,

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
			fmt.Println("You need to be logged in to view project details")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Create project service
		projectService := services.NewProjectService(queries)

		// Get the project
		project, err := projectService.GetProject(context.Background(), int32(projectID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving project: %v\n", err)
			return
		}

		// Display project details
		fmt.Println("Project Details:")
		fmt.Printf("ID: %d\n", project.ID)
		fmt.Printf("Name: %s\n", project.Name)

		if project.Description.Valid {
			fmt.Printf("Description: %s\n", project.Description.String)
		} else {
			fmt.Println("Description: None")
		}

		if project.Deadline.Valid {
			fmt.Printf("Deadline: %s\n", project.Deadline.Time.Format("2006-01-02"))
		} else {
			fmt.Println("Deadline: None")
		}

		fmt.Printf("Created: %s\n", project.CreatedAt.Time.Format("2006-01-02"))
		fmt.Printf("Last Updated: %s\n", project.UpdatedAt.Time.Format("2006-01-02"))

		// Get and display tasks associated with the project
		tasks, err := projectService.GetProjectTasks(context.Background(), int32(projectID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving project tasks: %v\n", err)
			return
		}

		// Count completed and pending tasks
		completedCount := 0
		pendingCount := 0
		for _, task := range tasks {
			if task.Status == "completed" {
				completedCount++
			} else {
				pendingCount++
			}
		}

		fmt.Println("\nTask Summary:")
		fmt.Printf("Total Tasks: %d\n", len(tasks))
		fmt.Printf("Completed: %d\n", completedCount)
		fmt.Printf("Pending: %d\n", pendingCount)

		if len(tasks) > 0 {
			fmt.Println("\nTasks:")
			for _, task := range tasks {
				statusStr := "[ ] "
				if task.Status == "completed" {
					statusStr = "[✓] "
				}
				fmt.Printf("%s#%d: %s\n", statusStr, task.ID, task.Description)
			}
		} else {
			fmt.Println("\nNo tasks associated with this project.")
			fmt.Println("Tip: Add a task to this project with: prod project task add " + args[0])
		}
	},
}

func init() {
	projectCmd.AddCommand(showProjectCmd)
}
