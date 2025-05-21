package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show [task_id]",
	Short: "Show detailed task information",
	Long: `Show detailed information about a specific task.
	
For example:
  prod task show 5  # Shows details for task with ID 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse task ID from arguments
		input, err := util.Input2Int(args[0])
		taskID, err := services.GetID(services.GetTaskMap, input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid task ID\n")
			return
		}

		// Initialize DB connection
		dbpool, err := util.InitDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer dbpool.Close()

		// Create queries and service
		queries := sqlc.New(dbpool)
		taskService := services.NewTaskService(queries)
		projectService := services.NewProjectService(queries)

		// Currently using a hardcoded user ID (1)
		// In a real app, you would get this from authentication
		userID := int32(1)

		// Get task details
		task, err := taskService.GetTask(context.Background(), int32(taskID), userID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to find task with ID %d: %v\n", taskID, err)
			return
		}

		// Get project name if task has a project
		var projectName string
		if task.ProjectID.Valid {
			project, err := projectService.GetProject(context.Background(), task.ProjectID.Int32, userID)
			if err == nil && project != nil {
				projectName = project.Name
			} else {
				projectName = fmt.Sprintf("ID %d", task.ProjectID.Int32)
			}
		}

		// Show task status with checkbox
		status := "[ ]"
		if task.CompletedAt.Valid {
			status = "[✓]"
		}

		// Show task details
		fmt.Printf("\n%s %s\n", status, task.Description)
		fmt.Printf("ID: %d\n", input)

		// Show priority if set
		if task.Priority.Valid {
			fmt.Printf("Priority: %s\n", task.Priority.String)
		}

		// Show project if set
		if task.ProjectID.Valid {
			fmt.Printf("Project: %s\n", projectName)
		}

		// Show dates
		if task.DueDate.Valid {
			fmt.Printf("Due: %s\n", task.DueDate.Time.Format("2006-01-02"))
		}
		if task.StartDate.Valid {
			fmt.Printf("Start: %s\n", task.StartDate.Time.Format("2006-01-02"))
		}
		if task.CompletedAt.Valid {
			fmt.Printf("Completed: %s\n", task.CompletedAt.Time.Format("2006-01-02"))
		}

		// Show recurrence if set
		if task.Recurrence.Valid && task.Recurrence.String != "" {
			fmt.Printf("Recurrence: %s\n", task.Recurrence.String)

			// Try to parse and display in a more readable format
			pattern, err := services.ParseRecurrence(task.Recurrence.String)
			if err == nil {
				fmt.Printf("Repeats: ")
				switch pattern.Type {
				case services.RecurrenceDaily:
					if pattern.Interval == 1 {
						fmt.Printf("Daily")
					} else {
						fmt.Printf("Every %d days", pattern.Interval)
					}
				case services.RecurrenceWeekly:
					if pattern.Interval == 1 {
						fmt.Printf("Weekly")
					} else {
						fmt.Printf("Every %d weeks", pattern.Interval)
					}
					if len(pattern.WeekDays) > 0 {
						// Convert weekday numbers to names
						weekdayNames := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
						var days []string
						for _, day := range pattern.WeekDays {
							if day >= 1 && day <= 7 {
								days = append(days, weekdayNames[day-1])
							}
						}
						fmt.Printf(" on %s", strings.Join(days, ", "))
					}
				case services.RecurrenceMonthly:
					if pattern.Interval == 1 {
						fmt.Printf("Monthly")
					} else {
						fmt.Printf("Every %d months", pattern.Interval)
					}
					if pattern.MonthDay > 0 {
						fmt.Printf(" on day %d", pattern.MonthDay)
					}
				case services.RecurrenceYearly:
					if pattern.Interval == 1 {
						fmt.Printf("Yearly")
					} else {
						fmt.Printf("Every %d years", pattern.Interval)
					}
					if pattern.YearlyDate != nil {
						monthNames := []string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
						if pattern.YearlyDate.Month >= 1 && pattern.YearlyDate.Month <= 12 {
							fmt.Printf(" on %s %d", monthNames[pattern.YearlyDate.Month-1], pattern.YearlyDate.Day)
						}
					}
				}
				fmt.Println()

				// Show until/count if set
				if pattern.Until != nil {
					fmt.Printf("Until: %s\n", pattern.Until.Format("2006-01-02"))
				}
				if pattern.Count > 0 {
					fmt.Printf("Occurrences: %d\n", pattern.Count)
				}

				// Show next occurrence if possible
				if task.Status != "completed" {
					var referenceDate time.Time
					if task.DueDate.Valid {
						referenceDate = task.DueDate.Time
					} else {
						referenceDate = time.Now()
					}
					nextDate, err := services.GetNextOccurrence(*pattern, referenceDate)
					if err == nil {
						fmt.Printf("Next occurrence: %s\n", nextDate.Format("2006-01-02"))
					}
				}
			}
		}

		// Show tags if any
		if len(task.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(task.Tags, ", "))
		} else {
			fmt.Printf("Tags: --\n")
		}

		fmt.Println()

		if task.Notes.Valid && task.Notes.String != "" {
			fmt.Printf("Notes: %s\n", task.Notes.String)
		}

		fmt.Printf("Created at: %s\n", task.CreatedAt.Time.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated at: %s\n", task.UpdatedAt.Time.Format("2006-01-02 15:04:05"))
	},
}

func init() {
	taskCmd.AddCommand(showCmd)
}
