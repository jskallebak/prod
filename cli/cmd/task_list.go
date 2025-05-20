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

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var (
	listPriority  string
	listProject   string
	tagsList      []string
	debugMode     bool
	showCompleted bool
	showAll       bool
	showToday     bool
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
		userService := services.NewUserService(queries)

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

		// Handle project flag
		if cmd.Flags().Changed("project") {
			projectPtr = &listProject
		} else {
			// if no project flag, checks for active project err == nil means there is a active project
			proj, err := userService.GetActiveProject(context.Background(), user.ID)
			if err == nil {
				str := strconv.Itoa(int(proj.ID))
				projectPtr = &str
			}
		}

		// Handle Today falg
		if cmd.Flags().Changed("today") {

		}

		if cmd.Flags().Changed("tag") {

		}

		// Status filter - by default only show pending and active tasks
		var status []string
		if showToday {
			status = []string{"pending", "active", "completed"}
		} else if !showCompleted && !showAll {
			status = []string{"pending", "active"}
		} else if showCompleted {
			status = []string{"completed"}
		} else if showAll {
			status = []string{"pending", "active"}
			projectPtr = nil
		}

		tasks, err := taskService.ListTasks(context.Background(), user.ID, priorityPtr, projectPtr, tagsList, status, showToday)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting list of tasks: %v\n", err)
			return
		}

		if showToday {
			now := time.Now()
			filtered := tasks[:0]
			for _, t := range tasks {
				fmt.Printf("DEBUG: Task %d status=%s due=%v\n", t.ID, t.Status, t.DueDate)
				if t.Status == "completed" {
					if t.DueDate.Valid {
						due := t.DueDate.Time
						if due.Year() == now.Year() && due.Month() == now.Month() && due.Day() == now.Day() {
							filtered = append(filtered, t)
						}
					}
				} else {
					filtered = append(filtered, t)
				}
			}
			tasks = filtered
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
			taskMap := map[int]int32{}
			err := services.MakeTaskMapFile(taskMap)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error making task map file: %v\n", err)
			}
			return
		}

		// Show title for the task list
		if showCompleted {
			fmt.Println("All tasks (including completed):")
		} else {
			fmt.Println("Pending tasks:")
		}

		// sl := SortTaskList(tasks)
		// fmt.Println(sl)

		taskMap := services.ProcessList(tasks, queries, user)
		err = services.MakeTaskMapFile(taskMap)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error making task map file: %v\n", err)
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

	listCmd.Flags().StringSliceVarP(&tagsList, "tags", "t", []string{}, "Filter tasks by project")

	// Add all flag
	listCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show tasks from all projects")

	// Add completed flag
	listCmd.Flags().BoolVarP(&showCompleted, "completed", "c", false, "Show completed tasks")

	// Add today flag
	listCmd.Flags().BoolVar(&showToday, "today", false, "Show tasks for today")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
