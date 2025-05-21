/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jskallebak/prod/internal/db/sqlc"
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
	showTable     bool
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

		if showTable {
			// Use a simple map for table view (ID to ID)
			taskMap := make(map[int]int32)
			for i, t := range tasks {
				taskMap[i+1] = t.ID
			}
			PrintTaskTableList(tasks, taskMap, queries, user)
			err = services.MakeTaskMapFile(taskMap)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error making task map file: %v\n", err)
			}
		} else {
			// Print and get the map from the multi-line function
			taskMap := PrintTaskMultiLineList(tasks, queries, user)
			err = services.MakeTaskMapFile(taskMap)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error making task map file: %v\n", err)
			}
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

	// Add table flag
	listCmd.Flags().BoolVarP(&showTable, "table", "T", false, "Show tasks in Taskwarrior-style table format")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// padAnsi will be moved to after the colors constants are defined in PrintTaskTableRow

func PrintTaskTableHeader() {
	fmt.Printf("%-4s %-40s %-6s %-12s %-15s %-12s %-12s %-20s\n",
		"ID", "Description", "Pri", "Due", "Tags", "Proj", "Status", "Completed")
}

// ansiRegexp to match ANSI color codes
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func PrintTaskTableRow(displayIdx int, task sqlc.Task, projectName string, altBg bool) {
	// ANSI colors
	const (
		reset        = "\033[0m"
		red          = "\033[31m"
		green        = "\033[32m"
		yellow       = "\033[33m"
		brightYellow = "\033[93m"
		brightGreen  = "\033[92m"
		brightCyan   = "\033[96m"
		darkGrayBg   = "\033[48;5;236m"
	)

	// Column widths (all inclusive of spacing)
	const (
		idWidth        = 5  // 4 + 1 space
		descWidth      = 41 // 40 + 1 space
		priWidth       = 7  // 6 + 1 space
		dueWidth       = 13 // 12 + 1 space
		tagsWidth      = 16 // 15 + 1 space
		projWidth      = 13 // 12 + 1 space
		statusWidth    = 13 // 12 + 1 space
		completedWidth = 20 // no trailing space needed
	)

	// Process the data
	id := fmt.Sprintf("%-*s", idWidth, fmt.Sprintf("%d", displayIdx))

	desc := task.Description
	if len([]rune(desc)) > descWidth-1 { // account for space
		desc = string([]rune(desc)[:descWidth-4]) + "..."
	}
	desc = fmt.Sprintf("%-*s", descWidth, desc)

	priority := "--"
	if task.Priority.Valid {
		priority = task.Priority.String
	}
	priority = fmt.Sprintf("%-*s", priWidth, priority)

	due := "--"
	overdue := false
	if task.DueDate.Valid {
		due = task.DueDate.Time.Format("2006-01-02")
		now := time.Now()

		// Compare only the dates, not times
		dueDate := time.Date(task.DueDate.Time.Year(), task.DueDate.Time.Month(), task.DueDate.Time.Day(), 0, 0, 0, 0, task.DueDate.Time.Location())
		todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

		if dueDate.Before(todayDate) && (!task.CompletedAt.Valid) {
			overdue = true
		}
	}
	due = fmt.Sprintf("%-*s", dueWidth, due)

	tags := "--"
	if len(task.Tags) > 0 {
		tags = strings.Join(task.Tags, ",")
	}
	tags = fmt.Sprintf("%-*s", tagsWidth, tags)

	proj := projectName
	if proj == "" {
		proj = "--"
	}
	proj = fmt.Sprintf("%-*s", projWidth, proj)

	status := task.Status
	if status == "" {
		status = "--"
	}
	status = fmt.Sprintf("%-*s", statusWidth, status)

	completed := "--"
	if task.CompletedAt.Valid {
		completed = task.CompletedAt.Time.Format("2006-01-02 15:04")
	}
	completed = fmt.Sprintf("%-*s", completedWidth, completed)

	// Now build the output with careful color control
	if altBg {
		// Print everything with a gray background
		fmt.Print(darkGrayBg)

		// ID (no color)
		fmt.Print(id)

		// Description (no color)
		fmt.Print(desc)

		// Priority (with color)
		if task.Priority.Valid {
			switch task.Priority.String {
			case "H":
				fmt.Print(red + priority[:len(priority)-1] + reset + darkGrayBg + " ")
			case "M":
				fmt.Print(yellow + priority[:len(priority)-1] + reset + darkGrayBg + " ")
			case "L":
				fmt.Print(green + priority[:len(priority)-1] + reset + darkGrayBg + " ")
			default:
				fmt.Print(priority)
			}
		} else {
			fmt.Print(priority)
		}

		// Due (with color for overdue)
		if overdue {
			fmt.Print(red + due[:len(due)-1] + reset + darkGrayBg + " ")
		} else {
			fmt.Print(due)
		}

		// Tags (no color)
		fmt.Print(tags)

		// Project (no color)
		fmt.Print(proj)

		// Status (with color)
		switch strings.TrimSpace(status) {
		case "completed":
			fmt.Print(brightGreen + strings.TrimSpace(status) + reset + darkGrayBg + strings.Repeat(" ", statusWidth-len(strings.TrimSpace(status))))
		case "pending":
			fmt.Print(brightYellow + strings.TrimSpace(status) + reset + darkGrayBg + strings.Repeat(" ", statusWidth-len(strings.TrimSpace(status))))
		case "active":
			fmt.Print(brightCyan + strings.TrimSpace(status) + reset + darkGrayBg + strings.Repeat(" ", statusWidth-len(strings.TrimSpace(status))))
		default:
			fmt.Print(status)
		}

		// Completed (colored if status is completed)
		if task.Status == "completed" && task.CompletedAt.Valid {
			fmt.Print(brightGreen + completed + reset)
		} else {
			fmt.Print(completed)
		}

		// End with erase to end of line and reset
		fmt.Print("\033[K" + reset + "\n")
	} else {
		// Regular row - apply colors directly to the specific fields

		// ID (no color)
		fmt.Print(id)

		// Description (no color)
		fmt.Print(desc)

		// Priority (with color)
		if task.Priority.Valid {
			switch task.Priority.String {
			case "H":
				fmt.Print(red + priority[:len(priority)-1] + reset + " ")
			case "M":
				fmt.Print(yellow + priority[:len(priority)-1] + reset + " ")
			case "L":
				fmt.Print(green + priority[:len(priority)-1] + reset + " ")
			default:
				fmt.Print(priority)
			}
		} else {
			fmt.Print(priority)
		}

		// Due (with color for overdue)
		if overdue {
			fmt.Print(red + due[:len(due)-1] + reset + " ")
		} else {
			fmt.Print(due)
		}

		// Tags (no color)
		fmt.Print(tags)

		// Project (no color)
		fmt.Print(proj)

		// Status (with color)
		switch strings.TrimSpace(status) {
		case "completed":
			fmt.Print(brightGreen + strings.TrimSpace(status) + reset + strings.Repeat(" ", statusWidth-len(strings.TrimSpace(status))))
		case "pending":
			fmt.Print(brightYellow + strings.TrimSpace(status) + reset + strings.Repeat(" ", statusWidth-len(strings.TrimSpace(status))))
		case "active":
			fmt.Print(brightCyan + strings.TrimSpace(status) + reset + strings.Repeat(" ", statusWidth-len(strings.TrimSpace(status))))
		default:
			fmt.Print(status)
		}

		// Completed (colored if status is completed)
		if task.Status == "completed" && task.CompletedAt.Valid {
			fmt.Print(brightGreen + completed + reset)
		} else {
			fmt.Print(completed)
		}

		fmt.Println()
	}
}

// PrintTaskTableList prints tasks in Taskwarrior-style table format
func PrintTaskTableList(tasks []sqlc.Task, taskMap map[int]int32, queries *sqlc.Queries, user *sqlc.User) {
	PrintTaskTableHeader()
	projectService := services.NewProjectService(queries)
	// Build a reverse map from task ID to display index
	idToDisplay := make(map[int32]int)
	for displayIdx, taskID := range taskMap {
		idToDisplay[taskID] = displayIdx
	}
	for i, t := range tasks {
		projectName := "--"
		if t.ProjectID.Valid {
			project, err := projectService.GetProject(context.Background(), t.ProjectID.Int32, user.ID)
			if err == nil && project != nil {
				projectName = project.Name
			}
		}
		displayIdx := idToDisplay[t.ID]
		altBg := (i%2 == 1)
		PrintTaskTableRow(displayIdx, t, projectName, altBg)
	}
}

// PrintTaskMultiLineList prints tasks in the multi-line icon-based format and returns the task map
func PrintTaskMultiLineList(tasks []sqlc.Task, queries *sqlc.Queries, user *sqlc.User) map[int]int32 {
	return services.ProcessList(tasks, queries, user)
}
