package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var (
	projectDescription string
	projectDeadline    string
)

var createProjectCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new project",
	Long: `Create a new project with the specified name.

Examples:
  prod project create "My New Project"
  prod project create "Work Project" --desc "Important work tasks" --deadline "2025-12-31"`,

	Args: cobra.ExactArgs(1),
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
			fmt.Println("You need to be logged in to create projects")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Get project name from argument
		projectName := args[0]

		// Create project service
		projectService := services.NewProjectService(queries)

		// Prepare params
		params := services.ProjectParams{
			Name: projectName,
		}

		// Add description if provided
		if cmd.Flags().Changed("desc") {
			params.Description = &projectDescription
		}

		// Add deadline if provided
		if cmd.Flags().Changed("deadline") {
			deadline, err := util.ParseDate(projectDeadline)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid deadline format: %v\n", err)
				fmt.Println("Use the format YYYY-MM-DD, e.g. 2025-12-31")
				return
			}
			params.Deadline = &deadline
		}

		// Create the project
		project, err := projectService.CreateProject(context.Background(), user.ID, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating project: %v\n", err)
			return
		}

		fmt.Println("Project created successfully!")
		fmt.Printf("ID: %d\n", project.ID)
		fmt.Printf("Name: %s\n", project.Name)

		if project.Description.Valid {
			fmt.Printf("Description: %s\n", project.Description.String)
		}

		if project.Deadline.Valid {
			fmt.Printf("Deadline: %s\n", project.Deadline.Time.Format("2006-01-02"))
		}

		fmt.Println()
		fmt.Println("To add tasks to this project:")
		fmt.Printf("  prod task create \"Task description\" --project %d\n", project.ID)
		fmt.Printf("  prod project task add %d [task-id]\n", project.ID)
	},
}

func init() {
	projectCmd.AddCommand(createProjectCmd)

	// Add flags
	createProjectCmd.Flags().StringVarP(&projectDescription, "desc", "d", "", "Project description")
	createProjectCmd.Flags().StringVarP(&projectDeadline, "deadline", "D", "", "Project deadline (YYYY-MM-DD)")
}
