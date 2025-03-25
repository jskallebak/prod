/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

const (
	ColorBlue   = "\033[34m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
	ColorReset  = "\033[0m"
)

// listCmd represents the list command
var (
	listPriority  string
	listProject   string
	debugMode     bool
	showCompleted bool
	showAll       bool
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
			uppercaseProject := strings.ToUpper(listProject)
			projectPtr = &uppercaseProject
		}

		// Status filter - by default only show pending and active tasks
		var status []string
		if !showCompleted && !showAll {
			status = []string{"pending", "active"}
		} else if showCompleted {
			status = []string{"completed"}
		} else if showAll {
			status = []string{"pending", "active", "completed"}
		}

		//TODO: remove when done with list.go
		if debugMode {
			fmt.Printf("Debug: Querying tasks for user ID: %d\n", user.ID)
			if priorityPtr != nil {
				fmt.Printf("Debug: Filtering by priority: %s\n", *priorityPtr)
			}
			if len(status) > 0 {
				fmt.Printf("Debug: Filtering by status: %s\n", status)
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

		tasks, err := taskService.ListTasks(context.Background(), user.ID, priorityPtr, projectPtr, status)
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

		// Show title for the task list
		if showCompleted {
			fmt.Println("All tasks (including completed):")
		} else {
			fmt.Println("Pending tasks:")
		}
		fmt.Println()

		for _, task := range tasks {
			// Show task status with checkbox
			status := "[ ]"
			if task.CompletedAt.Valid {
				status = "[✓]"
			}
			coloredDescription := fmt.Sprintf("%s%s%s", ColorGreen, task.Description, ColorReset)
			fmt.Printf("%s #%d %s\n", status, task.ID, coloredDescription)

			// Show task details with emojis and consistent formatting
			fmt.Printf("    🎯\tStatus: %s\n", task.Status) // Target/goal
			if task.CompletedAt.Valid {
				fmt.Printf("    ✅\tCompleted: %s\n", task.CompletedAt.Time.Format("2006-01-02 15:04"))
			}
			if task.Priority.Valid {
				// Convert priority letter to full name
				priorityName := "Unknown"
				switch task.Priority.String {
				case "H":
					priorityName = ColorRed + "High" + ColorReset
				case "M":
					priorityName = ColorYellow + "Medium" + ColorReset
				case "L":
					priorityName = ColorGreen + "Low" + ColorReset
				}
				fmt.Printf("    🔄\tPriority: %s\n", priorityName)
			} else {
				fmt.Printf("    🔄\tPriority: --\n")
			}
			if task.ProjectID.Valid {
				// Get project name instead of just showing the ID
				projectService := services.NewProjectService(queries)
				project, err := projectService.GetProject(context.Background(), task.ProjectID.Int32, user.ID)
				if err == nil && project != nil {
					fmt.Printf("    📁\tProject: %s\n", project.Name)
				} else {
					fmt.Printf("    📁\tProject: ID %d\n", task.ProjectID.Int32)
				}
			} else {
				fmt.Printf("    📁\tProject: --\n")
			}
			if task.StartDate.Valid {
				fmt.Printf("    📅\tStarted at: %s\n", task.StartDate.Time.Format("Mon, Jan 2, 2006"))
			} else {
				fmt.Printf("    📅\tStarted at: --\n")
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
		}
	},
}

func init() {
	taskCmd.AddCommand(listCmd)

	// Add priority flag for filtering tasks
	listCmd.Flags().StringVarP(&listPriority, "priority", "p", "", "Filter tasks by priority (H, M, L)")

	// Add debug flag
	listCmd.Flags().BoolVar(&debugMode, "debug", false, "Enable debug mode")

	// Add project flag
	listCmd.Flags().StringVarP(&listProject, "project", "P", "", "Filter tasks by project")

	// Add completed flag
	listCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all tasks (including completed)")

	listCmd.Flags().BoolVarP(&showCompleted, "completed", "c", false, "Show completed tasks")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
