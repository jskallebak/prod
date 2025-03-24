package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var listProjectCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long: `List all your projects.

Example:
  prod project list

This command shows all your projects with their ID, name, description, and deadline (if set).`,

	Run: func(cmd *cobra.Command, args []string) {
		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		// Get authenticated user
		authService := services.NewAuthService(queries)
		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Println("You need to be logged in to list projects")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Create project service
		projectService := services.NewProjectService(queries)

		// List the projects
		projects, err := projectService.ListProjects(context.Background(), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing projects: %v\n", err)
			return
		}

		if len(projects) == 0 {
			fmt.Println("No projects found")
			fmt.Println("\nTip: Create a project with: prod project create \"My Project\"")
			return
		}

		fmt.Printf("Your projects (%d):\n\n", len(projects))

		for _, project := range projects {
			fmt.Printf("ID: %d\n", project.ID)
			fmt.Printf("Name: %s\n", project.Name)

			if project.Description.Valid {
				fmt.Printf("Description: %s\n", project.Description.String)
			}

			if project.Deadline.Valid {
				fmt.Printf("Deadline: %s\n", project.Deadline.Time.Format("2006-01-02"))
			}

			// Display creation date
			fmt.Printf("Created: %s\n", project.CreatedAt.Time.Format("2006-01-02"))
			fmt.Println()
		}
	},
}

func init() {
	projectCmd.AddCommand(listProjectCmd)
}
