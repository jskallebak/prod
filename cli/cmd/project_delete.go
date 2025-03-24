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

var forceDeleteProject bool

var deleteProjectCmd = &cobra.Command{
	Use:   "delete [project-id]",
	Short: "Delete a project",
	Long: `Delete a project and optionally all its associated tasks.

Examples:
  prod project delete 1
  prod project delete 1 --force  # Skip confirmation prompt`,

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
			fmt.Println("You need to be logged in to delete projects")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Create project service
		projectService := services.NewProjectService(queries)

		// Get project to confirm deletion
		project, err := projectService.GetProject(context.Background(), int32(projectID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving project: %v\n", err)
			return
		}

		// Get project tasks to inform user what will be affected
		tasks, err := projectService.GetProjectTasks(context.Background(), int32(projectID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving project tasks: %v\n", err)
			return
		}

		// Confirm deletion unless --force is used
		if !forceDeleteProject {
			fmt.Printf("You are about to delete project: %s (ID: %d)\n", project.Name, project.ID)

			if len(tasks) > 0 {
				fmt.Printf("This project has %d associated tasks. ", len(tasks))
				fmt.Println("Tasks will be removed from the project but not deleted.")
			}

			fmt.Print("Are you sure you want to proceed? (y/N): ")
			var answer string
			fmt.Scanln(&answer)

			if answer != "y" && answer != "Y" {
				fmt.Println("Operation cancelled")
				return
			}
		}

		// Delete the project
		err = projectService.DeleteProject(context.Background(), int32(projectID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting project: %v\n", err)
			return
		}

		fmt.Printf("Project '%s' (ID: %d) deleted successfully\n", project.Name, project.ID)
	},
}

func init() {
	projectCmd.AddCommand(deleteProjectCmd)

	// Add force flag for skipping confirmation
	deleteProjectCmd.Flags().BoolVarP(&forceDeleteProject, "force", "f", false, "Skip confirmation prompt")
}
