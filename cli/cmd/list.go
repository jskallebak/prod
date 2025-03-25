/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var (
	listPriority      string
	listProject       string
	listProjectSelect bool
	debugMode         bool
	showCompleted     bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your tasks",
	Long: `List all your tasks or filter them by priority.

Examples:
  prod task list                 # List all incomplete tasks
  prod task list --completed     # List all tasks including completed ones
  prod task list --priority=H    # List only high priority tasks
  prod task list -p M            # List only medium priority tasks
  
Priority levels:
  H - High
  M - Medium
  L - Low
  
  prod task list --project=ProjectName
  prod task list -P ProjectName`,

	Run: func(cmd *cobra.Command, args []string) {
		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		taskService := services.NewTaskService(queries)
		authService := services.NewAuthService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Println("Needs to be logged in to show tasks")
			return
		}

		// Handle priority flag
		var priorityPtr *string
		if cmd.Flags().Changed("priority") {
			// Convert priority to uppercase for case-insensitive matching
			uppercasePriority := strings.ToUpper(listPriority)
			priorityPtr = &uppercasePriority
		}

		var projectPtr *string
		if cmd.Flags().Changed("project") {
			// Project ID or name was directly provided
			// First, try to interpret as ID
			projectID, err := strconv.Atoi(listProject)
			if err == nil {
				// It's a numeric ID
				projectIDStr := fmt.Sprintf("%d", projectID)
				projectPtr = &projectIDStr
			} else {
				// It's likely a project name - search for matching project
				projectService := services.NewProjectService(queries)
				projects, err := projectService.ListProjects(context.Background(), user.ID)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error listing projects: %v\n", err)
					return
				}

				// Search for project by name (case-insensitive)
				found := false
				searchName := strings.ToLower(listProject)
				for _, project := range projects {
					if strings.Contains(strings.ToLower(project.Name), searchName) {
						projectIDStr := fmt.Sprintf("%d", project.ID)
						projectPtr = &projectIDStr
						found = true
						break
					}
				}

				if !found {
					fmt.Printf("No project found matching '%s'\n", listProject)
					return
				}
			}
		} else if listProjectSelect {
			// User wants to select a project interactively
			projectService := services.NewProjectService(queries)
			projects, err := projectService.ListProjects(context.Background(), user.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing projects: %v\n", err)
				return
			}

			if len(projects) == 0 {
				fmt.Println("No projects found. Create a project first with: prod project create \"Project Name\"")
				return
			}

			// Display list of projects
			fmt.Println("Available projects:")
			// Create a map to store projects by ID for easy lookup
			projectsById := make(map[int]sqlc.Project)
			projectIDs := make([]int, 0, len(projects))

			for _, project := range projects {
				projectsById[int(project.ID)] = project
				projectIDs = append(projectIDs, int(project.ID))
			}

			// Display projects with their actual IDs
			for _, id := range projectIDs {
				project := projectsById[id]
				fmt.Printf("%d. %s\n", project.ID, project.Name)
			}

			// Prompt for selection
			var selection int
			fmt.Print("\nSelect a project by ID: ")
			_, err = fmt.Scanln(&selection)

			// Validate the selection is a valid project ID
			_, valid := projectsById[selection]
			if err != nil || !valid {
				fmt.Fprintf(os.Stderr, "Invalid project ID. Please enter one of the listed project IDs\n")
				return
			}

			// Convert to string and set as project pointer
			selectedProject := fmt.Sprintf("%d", selection)
			projectPtr = &selectedProject
		}

		// Status filter - by default only show pending tasks
		var statusPtr *string
		if !showCompleted {
			status := "pending"
			statusPtr = &status
		}

		//TODO: remove when done with list.go
		if debugMode {
			fmt.Printf("Debug: Querying tasks for user ID: %d\n", user.ID)
			if priorityPtr != nil {
				fmt.Printf("Debug: Filtering by priority: %s\n", *priorityPtr)
			}
			if statusPtr != nil {
				fmt.Printf("Debug: Filtering by status: %s\n", *statusPtr)
			}

			// Print DB connection info
			fmt.Printf("Debug: Database connection established: %v\n", dbpool != nil)

			// Print SQL query test
			countTest, err := queries.CountTasks(context.Background(), sqlc.CountTasksParams{
				UserID: pgtype.Int4{
					Int32: int32(user.ID),
					Valid: true,
				},
			})
			if err != nil {
				fmt.Printf("Debug: Error running test query: %v\n", err)
			} else {
				fmt.Printf("Debug: Database has %d total tasks, %d pending, %d completed\n",
					countTest.TotalTasks, countTest.PendingTasks, countTest.CompletedTasks)
			}
		}

		tasks, err := taskService.ListTasks(context.Background(), user.ID, priorityPtr, projectPtr, statusPtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting list of tasks: %v\n", err)
			return
		}

		if debugMode {
			fmt.Printf("Debug: Query returned %d tasks\n", len(tasks))
		}

		if len(tasks) == 0 {
			if cmd.Flags().Changed("priority") {
				fmt.Printf("No tasks found with priority '%s'\n", listPriority)
			} else if cmd.Flags().Changed("project") {
				fmt.Printf("No tasks found with project '%s'\n", listProject)
			} else if !showCompleted {
				fmt.Println("No pending tasks found")
				fmt.Println("\nTip: Create a task with: prod task add \"My first task\"")
				fmt.Println("     To see completed tasks, use: prod task list --completed")
			} else {
				fmt.Println("No tasks found")
				fmt.Println("\nTip: Create a task with: prod task add \"My first task\"")
			}
			return
		}

		// Reverse the order of tasks (newest first)
		for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
			tasks[i], tasks[j] = tasks[j], tasks[i]
		}

		// Show title for the task list
		if showCompleted {
			fmt.Println("🗒️  All tasks (including completed):")
		} else {
			fmt.Println("🗒️  Pending tasks:")
		}
		fmt.Println("----------------------------------------------")

		for i, task := range tasks {
			// Create status indicator
			statusSymbol := "[ ]"
			if task.CompletedAt.Valid {
				statusSymbol = "[✓]"
			}

			// Print task header with ID and status (without priority)
			fmt.Printf("%s #%d %s\n", statusSymbol, task.ID, task.Description)

			// Print priority on a new line, showing "None" if not set
			if task.Priority.Valid {
				fmt.Printf("    🔄\tPriority: %s\n", task.Priority.String)
			} else {
				fmt.Printf("    🔄\tPriority: None\n")
			}

			// Print metadata indented
			if task.CompletedAt.Valid {
				fmt.Printf("    ✅\tCompleted: %s\n", task.CompletedAt.Time.Format("2006-01-02 15:04"))
			}

			// Print project info, showing "None" if not set
			if task.ProjectID.Valid {
				projectService := services.NewProjectService(queries)
				project, err := projectService.GetProject(context.Background(), task.ProjectID.Int32, user.ID)
				if err == nil && project != nil {
					fmt.Printf("    📁\tProject: %s\n", project.Name)
				} else {
					fmt.Printf("    📁\tProject ID: %d\n", task.ProjectID.Int32)
				}
			} else {
				fmt.Printf("    📁\tProject: None\n")
			}

			// Print due date if exists, or "None" if not set
			if task.DueDate.Valid {
				// Calculate if the task is overdue
				dueStatus := ""
				if task.DueDate.Time.Before(time.Now()) && !task.CompletedAt.Valid {
					dueStatus = " (OVERDUE)"
				}
				fmt.Printf("    📅\tDue: %s%s\n", task.DueDate.Time.Format("Mon, Jan 2, 2006"), dueStatus)
			} else {
				fmt.Printf("    📅\tDue: None\n")
			}

			// Print tags if exists
			if len(task.Tags) > 0 {
				fmt.Printf("    🏷️\tTags: %s\n", strings.Join(task.Tags, ", "))
			}

			// Add separator between tasks, except after the last one
			if i < len(tasks)-1 {
				fmt.Println("----------------------------------------------")
			}
		}
		fmt.Println("----------------------------------------------")
		fmt.Printf("Total: %d tasks\n", len(tasks))
	},
}

func init() {
	taskCmd.AddCommand(listCmd)

	// Add priority flag for filtering tasks
	listCmd.Flags().StringVarP(&listPriority, "priority", "p", "", "Filter tasks by priority (H, M, L)")

	// Add debug flag
	listCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug mode")

	// Add project flag
	listCmd.Flags().StringVarP(&listProject, "project", "P", "", "Filter tasks by project ID or name")

	// Add project selection flag
	listCmd.Flags().BoolVarP(&listProjectSelect, "select-project", "s", false, "Select project interactively")

	// Add completed flag
	listCmd.Flags().BoolVarP(&showCompleted, "completed", "c", false, "Show completed tasks")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
