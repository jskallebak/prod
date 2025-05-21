package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var (
	recurType       string
	recurInterval   int
	recurWeekDays   []string
	recurMonthDay   int
	recurUntil      string
	recurCount      int
	clearRecurrence bool
)

// recurCmd represents the recur command
var recurCmd = &cobra.Command{
	Use:   "recur [task_id]",
	Short: "Set a task to recur on a schedule",
	Long: `Set a task to recur on a specified schedule.
	
Examples:
  prod task recur 5 --type=daily
  prod task recur 5 --type=weekly --interval=2 --weekdays=mon,wed,fri
  prod task recur 5 --type=monthly --monthday=15
  prod task recur 5 --type=yearly --count=10
  prod task recur 5 --type=daily --until=2026-01-01`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "Error: Task ID is required\n")
			cmd.Help()
			return
		}

		// Parse task ID from argument
		input, err := util.Input2Int(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid task ID format\n")
			return
		}

		// Initialize DB connection
		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			fmt.Fprintf(os.Stderr, "Error connecting to database\n")
			return
		}
		defer dbpool.Close()

		// Create task service and get authenticated user
		taskService := services.NewTaskService(queries)
		authService := services.NewAuthService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: You need to be logged in to set task recurrence\n")
			fmt.Fprintf(os.Stderr, "Use 'prod login' to authenticate\n")
			return
		}

		// Convert CLI task ID to database ID
		taskID, err := services.GetID(services.GetTaskMap, input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		// Get the task to verify it exists and belongs to the user
		task, err := taskService.GetTask(context.Background(), taskID, user.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to find task with ID %d: %v\n", input, err)
			return
		}

		// Print current task info
		fmt.Printf("Task: %s\n", task.Description)
		if task.Recurrence.Valid && task.Recurrence.String != "" {
			fmt.Printf("Current recurrence: %s\n", task.Recurrence.String)
		}

		// If clear flag is set, remove recurrence
		if clearRecurrence {
			updatedTask, err := taskService.UpdateTaskRecurrence(context.Background(), taskID, user.ID, "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error clearing recurrence: %v\n", err)
				return
			}

			fmt.Printf("Recurrence removed from task %d\n", input)
			fmt.Printf("Description: %s\n", updatedTask.Description)
			return
		}

		// If we got here, we're setting a recurrence pattern
		// Build recurrence pattern string
		recurrencePattern, err := buildRecurrencePattern(cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}

		// Update the task with the recurrence pattern
		updatedTask, err := taskService.UpdateTaskRecurrence(context.Background(), taskID, user.ID, recurrencePattern)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error setting recurrence: %v\n", err)
			return
		}

		fmt.Printf("Task %d set to recur %s\n", input, recurrencePattern)
		fmt.Printf("Description: %s\n", updatedTask.Description)

		// Print the next occurrence date if possible
		pattern, err := services.ParseRecurrence(recurrencePattern)
		if err == nil {
			var referenceDate time.Time
			if updatedTask.DueDate.Valid {
				referenceDate = updatedTask.DueDate.Time
			} else {
				referenceDate = time.Now()
			}

			nextDate, err := services.GetNextOccurrence(*pattern, referenceDate)
			if err == nil {
				fmt.Printf("Next occurrence: %s\n", nextDate.Format("2006-01-02"))
			}
		}
	},
}

// buildRecurrencePattern constructs a recurrence pattern string from command flags
func buildRecurrencePattern(cmd *cobra.Command) (string, error) {
	// Validate frequency type
	if !cmd.Flags().Changed("type") {
		return "", fmt.Errorf("recurrence type is required")
	}

	recurType = strings.ToLower(recurType)
	switch recurType {
	case "daily", "weekly", "monthly", "yearly":
		// Valid type
	default:
		return "", fmt.Errorf("invalid recurrence type: must be daily, weekly, monthly, or yearly")
	}

	// Start building the pattern
	pattern := recurType

	// Add interval if specified
	if cmd.Flags().Changed("interval") {
		if recurInterval < 1 {
			return "", fmt.Errorf("interval must be at least 1")
		}
		pattern = fmt.Sprintf("%s:%d", pattern, recurInterval)
	} else {
		pattern = fmt.Sprintf("%s:1", pattern) // Default interval is 1
	}

	// Add frequency-specific details
	switch recurType {
	case "weekly":
		if cmd.Flags().Changed("weekdays") {
			weekdayMap := map[string]int{
				"mon": 1, "monday": 1,
				"tue": 2, "tuesday": 2,
				"wed": 3, "wednesday": 3,
				"thu": 4, "thursday": 4,
				"fri": 5, "friday": 5,
				"sat": 6, "saturday": 6,
				"sun": 7, "sunday": 7,
			}

			// Convert weekday names to numbers
			var weekdayNums []int
			for _, day := range recurWeekDays {
				day = strings.ToLower(day)
				if num, ok := weekdayMap[day]; ok {
					weekdayNums = append(weekdayNums, num)
				} else {
					return "", fmt.Errorf("invalid weekday: %s", day)
				}
			}

			if len(weekdayNums) == 0 {
				// Default to all weekdays if none specified
				pattern = fmt.Sprintf("%s:", pattern)
			} else {
				weekdaysStr := make([]string, len(weekdayNums))
				for i, num := range weekdayNums {
					weekdaysStr[i] = fmt.Sprintf("%d", num)
				}
				pattern = fmt.Sprintf("%s:%s", pattern, strings.Join(weekdaysStr, ","))
			}
		} else {
			pattern = fmt.Sprintf("%s:", pattern) // No specific weekdays
		}

	case "monthly":
		if cmd.Flags().Changed("monthday") {
			if recurMonthDay < 1 || recurMonthDay > 31 {
				return "", fmt.Errorf("month day must be between 1 and 31")
			}
			pattern = fmt.Sprintf("%s:%d", pattern, recurMonthDay)
		} else {
			pattern = fmt.Sprintf("%s:", pattern) // No specific month day
		}

	case "yearly":
		// For yearly, we'd need to specify a month and day, but for simplicity
		// we'll just use the task's due date as the yearly recurrence date
		pattern = fmt.Sprintf("%s:", pattern)
	}

	// Add count limit if specified
	if cmd.Flags().Changed("count") {
		if recurCount < 1 {
			return "", fmt.Errorf("count must be at least 1")
		}
		pattern = fmt.Sprintf("%s:count:%d", pattern, recurCount)
	}

	// Add until date if specified
	if cmd.Flags().Changed("until") {
		_, err := time.Parse("2006-01-02", recurUntil)
		if err != nil {
			return "", fmt.Errorf("invalid until date format (use YYYY-MM-DD)")
		}
		pattern = fmt.Sprintf("%s:until:%s", pattern, recurUntil)
	}

	return pattern, nil
}

func init() {
	taskCmd.AddCommand(recurCmd)

	// Define flags for recurrence command
	recurCmd.Flags().StringVarP(&recurType, "type", "t", "", "Recurrence type (daily, weekly, monthly, yearly)")
	recurCmd.Flags().IntVarP(&recurInterval, "interval", "i", 1, "Recurrence interval (e.g., every 2 days)")
	recurCmd.Flags().StringSliceVarP(&recurWeekDays, "weekdays", "w", []string{}, "Days of week for weekly recurrence (mon,tue,wed,thu,fri,sat,sun)")
	recurCmd.Flags().IntVarP(&recurMonthDay, "monthday", "m", 0, "Day of month for monthly recurrence (1-31)")
	recurCmd.Flags().StringVarP(&recurUntil, "until", "u", "", "Recur until date (YYYY-MM-DD)")
	recurCmd.Flags().IntVarP(&recurCount, "count", "c", 0, "Recur this many times")
	recurCmd.Flags().BoolVarP(&clearRecurrence, "clear", "C", false, "Clear recurrence pattern from task")
}
