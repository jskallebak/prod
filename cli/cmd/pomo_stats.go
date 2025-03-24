package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var (
	statsTimeFrame string
)

var statsCmd = &cobra.Command{
	Use:   "stats [task-id]",
	Short: "Show Pomodoro statistics",
	Long: `Display statistics for your Pomodoro sessions, optionally filtered by task.

Examples:
  prod pomo stats            # Show overall Pomodoro statistics
  prod pomo stats 5          # Show Pomodoro statistics for task with ID 5
  prod pomo stats --day      # Show Pomodoro statistics for today
  prod pomo stats --week     # Show Pomodoro statistics for this week
  prod pomo stats --month    # Show Pomodoro statistics for this month`,

	Args: cobra.MaximumNArgs(1),
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
			fmt.Println("You need to be logged in to view Pomodoro statistics")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Parse task ID if provided
		var taskID *int32
		if len(args) == 1 {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid task ID: %v\n", err)
				return
			}

			// Verify the task exists and belongs to the user
			taskService := services.NewTaskService(queries)
			task, err := taskService.GetTask(context.Background(), int32(id), user.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return
			}

			intID := int32(id)
			taskID = &intID

			fmt.Printf("Pomodoro Statistics for Task: %s (ID: %d)\n\n", task.Description, task.ID)
		} else {
			fmt.Println("Overall Pomodoro Statistics")
		}

		// Determine time range
		var startDate, endDate *time.Time
		now := time.Now()

		if statsTimeFrame == "day" {
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			startDate = &startOfDay
			fmt.Printf("Time Frame: Today (%s)\n\n", startOfDay.Format("2006-01-02"))
		} else if statsTimeFrame == "week" {
			// Get start of week (Sunday)
			daysToSunday := int(now.Weekday())
			startOfWeek := now.AddDate(0, 0, -daysToSunday)
			startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, now.Location())
			startDate = &startOfWeek
			fmt.Printf("Time Frame: This Week (from %s)\n\n", startOfWeek.Format("2006-01-02"))
		} else if statsTimeFrame == "month" {
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			startDate = &startOfMonth
			fmt.Printf("Time Frame: This Month (%s)\n\n", startOfMonth.Format("2006-01"))
		} else {
			fmt.Println("Time Frame: All Time\n")
		}

		// Get statistics
		pomoService := services.NewPomodoroService(queries)
		stats, err := pomoService.GetSessionStats(context.Background(), user.ID, taskID, startDate, endDate)
		if err != nil {
			fmt.Printf("Error retrieving Pomodoro statistics: %v\n", err)
			return
		}

		// Display statistics
		totalSessions := stats["total_sessions"].(int32)
		if totalSessions == 0 {
			fmt.Println("No Pomodoro sessions found for the selected criteria")
			return
		}

		fmt.Printf("Total Sessions: %d\n", totalSessions)
		fmt.Printf("Completed: %d\n", stats["completed_sessions"].(int32))
		fmt.Printf("Cancelled: %d\n", stats["cancelled_sessions"].(int32))

		completionRate := float64(stats["completed_sessions"].(int32)) / float64(totalSessions) * 100
		fmt.Printf("Completion Rate: %.1f%%\n\n", completionRate)

		fmt.Printf("Total Work Time: %d minutes\n", stats["total_work_mins"].(int32))
		fmt.Printf("Total Break Time: %d minutes\n", stats["total_break_mins"].(int32))
		fmt.Printf("Total Time: %d minutes\n", stats["total_duration_mins"].(int32))
		fmt.Printf("Average Session: %.1f minutes\n\n", float64(stats["avg_duration_mins"].(float64)))

		// Show most productive day/hour if available
		if mpd, ok := stats["most_productive_day"].(time.Time); ok && !mpd.IsZero() {
			fmt.Printf("Most Productive Day: %s\n", mpd.Format("2006-01-02"))
		}

		if mph, ok := stats["most_productive_hour"].(int32); ok && mph >= 0 {
			var ampm string
			hour := mph
			if hour < 12 {
				ampm = "AM"
			} else {
				ampm = "PM"
			}
			if hour > 12 {
				hour -= 12
			}
			if hour == 0 {
				hour = 12
			}
			fmt.Printf("Most Productive Hour: %d:00 %s\n", hour, ampm)
		}
	},
}

func init() {
	pomoCmd.AddCommand(statsCmd)

	// Add flags
	statsCmd.Flags().StringVar(&statsTimeFrame, "time", "", "Time frame for statistics (day, week, month)")
	statsCmd.Flags().BoolP("day", "d", false, "Show statistics for today")
	statsCmd.Flags().BoolP("week", "w", false, "Show statistics for this week")
	statsCmd.Flags().BoolP("month", "m", false, "Show statistics for this month")

	// Bind bool flags to the time frame string
	statsCmd.PreRun = func(cmd *cobra.Command, args []string) {
		dayFlag, _ := cmd.Flags().GetBool("day")
		weekFlag, _ := cmd.Flags().GetBool("week")
		monthFlag, _ := cmd.Flags().GetBool("month")

		if dayFlag {
			statsTimeFrame = "day"
		} else if weekFlag {
			statsTimeFrame = "week"
		} else if monthFlag {
			statsTimeFrame = "month"
		}
	}
}
