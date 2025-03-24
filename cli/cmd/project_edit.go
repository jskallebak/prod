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

var (
	editProjectName        string
	editProjectDescription string
	editProjectDeadline    string
	clearDescription       bool
	clearDeadline          bool
)

var editProjectCmd = &cobra.Command{
	Use:   "edit [project-id]",
	Short: "Edit project details",
	Long: `Edit details of an existing project.

Examples:
  prod project edit 1 --name "Updated Project Name"
  prod project edit 1 --desc "New description" --deadline "2025-06-30"
  prod project edit 1 --clear-desc  # Remove description
  prod project edit 1 --clear-deadline  # Remove deadline

At least one flag must be specified.`,

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Ensure at least one edit flag is specified
		if !cmd.Flags().Changed("name") &&
			!cmd.Flags().Changed("desc") &&
			!cmd.Flags().Changed("deadline") &&
			!clearDescription &&
			!clearDeadline {
			fmt.Println("Error: At least one edit flag must be specified")
			fmt.Println("Use --help for more information")
			return
		}

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
			fmt.Println("You need to be logged in to edit projects")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Create project service
		projectService := services.NewProjectService(queries)

		// Get current project to show changes
		currentProject, err := projectService.GetProject(context.Background(), int32(projectID), user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving project: %v\n", err)
			return
		}

		// Prepare update parameters
		params := services.ProjectParams{}

		// Set name if provided
		if cmd.Flags().Changed("name") {
			params.Name = editProjectName
		} else {
			// Keep the current name
			params.Name = currentProject.Name
		}

		// Handle description updates/clearing
		if clearDescription {
			// Set empty description
			emptyDesc := ""
			params.Description = &emptyDesc
		} else if cmd.Flags().Changed("desc") {
			params.Description = &editProjectDescription
		}

		// Handle deadline updates/clearing
		if clearDeadline {
			// We'll use nil deadline to clear it in the service layer
			// This is handled specially in the project service
		} else if cmd.Flags().Changed("deadline") {
			deadline, err := util.ParseDate(editProjectDeadline)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid deadline format: %v\n", err)
				fmt.Println("Use the format YYYY-MM-DD, e.g. 2025-12-31")
				return
			}
			params.Deadline = &deadline
		}

		// Update the project
		updatedProject, err := projectService.UpdateProject(context.Background(), int32(projectID), user.ID, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating project: %v\n", err)
			return
		}

		fmt.Println("Project updated successfully!")
		fmt.Printf("ID: %d\n", updatedProject.ID)
		fmt.Printf("Name: %s\n", updatedProject.Name)

		if updatedProject.Description.Valid {
			fmt.Printf("Description: %s\n", updatedProject.Description.String)
		} else {
			fmt.Println("Description: None")
		}

		if updatedProject.Deadline.Valid {
			fmt.Printf("Deadline: %s\n", updatedProject.Deadline.Time.Format("2006-01-02"))
		} else {
			fmt.Println("Deadline: None")
		}
	},
}

func init() {
	projectCmd.AddCommand(editProjectCmd)

	// Add flags for editing
	editProjectCmd.Flags().StringVarP(&editProjectName, "name", "n", "", "New project name")
	editProjectCmd.Flags().StringVarP(&editProjectDescription, "desc", "d", "", "New project description")
	editProjectCmd.Flags().StringVarP(&editProjectDeadline, "deadline", "D", "", "New project deadline (YYYY-MM-DD)")

	// Add flags for clearing fields
	editProjectCmd.Flags().BoolVar(&clearDescription, "clear-desc", false, "Clear project description")
	editProjectCmd.Flags().BoolVar(&clearDeadline, "clear-deadline", false, "Clear project deadline")
}
