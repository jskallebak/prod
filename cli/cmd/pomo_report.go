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
	reportPeriod string
	reportFormat string
	reportOutput string
)

var reportCmd = &cobra.Command{
	Use:   "report [task-id]",
	Short: "Generate Pomodoro reports",
	Long: `Generate reports on your Pomodoro usage, optionally filtered by task.

Examples:
  prod pomo report            # Generate a report for all Pomodoros
  prod pomo report 5          # Generate a report for task with ID 5
  prod pomo report --period week    # Report for this week
  prod pomo report --period month   # Report for this month
  prod pomo report --format text    # Plain text report
  prod pomo report --output file    # Output to file`,

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
			fmt.Println("You need to be logged in to generate Pomodoro reports")
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

			fmt.Printf("Pomodoro Report for Task: %s (ID: %d)\n\n", task.Description, task.ID)
		} else {
			fmt.Println("Pomodoro Report - All Tasks")
		}

		// Determine time range for report
		var startDate, endDate *time.Time
		now := time.Now()

		switch reportPeriod {
		case "day":
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			startDate = &startOfDay
			fmt.Printf("Period: Today (%s)\n\n", startOfDay.Format("2006-01-02"))
		case "week":
			// Get start of week (Sunday)
			daysToSunday := int(now.Weekday())
			startOfWeek := now.AddDate(0, 0, -daysToSunday)
			startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, now.Location())
			startDate = &startOfWeek
			fmt.Printf("Period: This Week (from %s)\n\n", startOfWeek.Format("2006-01-02"))
		case "month":
			startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			startDate = &startOfMonth
			fmt.Printf("Period: This Month (%s)\n\n", startOfMonth.Format("2006-01"))
		case "year":
			startOfYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
			startDate = &startOfYear
			fmt.Printf("Period: This Year (%d)\n\n", now.Year())
		default:
			fmt.Println("Period: All Time\n")
		}

		// Get pomodoro service
		pomoService := services.NewPomodoroService(queries)

		// Generate report
		report, err := pomoService.GenerateReport(context.Background(), user.ID, taskID, startDate, endDate)
		if err != nil {
			fmt.Printf("Error generating Pomodoro report: %v\n", err)
			return
		}

		// Check if report has any data
		if report.TotalSessions == 0 {
			fmt.Println("No Pomodoro sessions found for the selected criteria")
			return
		}

		// Display report stats
		fmt.Printf("Total Sessions: %d\n", report.TotalSessions)
		fmt.Printf("Completed: %d (%.1f%%)\n",
			report.CompletedSessions,
			float64(report.CompletedSessions)/float64(report.TotalSessions)*100)
		fmt.Printf("Cancelled: %d (%.1f%%)\n",
			report.CancelledSessions,
			float64(report.CancelledSessions)/float64(report.TotalSessions)*100)

		// Time stats
		fmt.Printf("\nTotal Time: %s\n", formatDurationSeconds(report.TotalTimeSeconds))
		fmt.Printf("Work Time: %s (%.1f%%)\n",
			formatDurationSeconds(report.WorkTimeSeconds),
			float64(report.WorkTimeSeconds)/float64(report.TotalTimeSeconds)*100)
		fmt.Printf("Break Time: %s (%.1f%%)\n",
			formatDurationSeconds(report.BreakTimeSeconds),
			float64(report.BreakTimeSeconds)/float64(report.TotalTimeSeconds)*100)
		fmt.Printf("Pause Time: %s (%.1f%%)\n",
			formatDurationSeconds(report.PauseTimeSeconds),
			float64(report.PauseTimeSeconds)/float64(report.TotalTimeSeconds)*100)

		// Productivity stats
		fmt.Printf("\nAverage Session: %s\n", formatDurationSeconds(report.AvgSessionSeconds))
		fmt.Printf("Average Work Session: %s\n", formatDurationSeconds(report.AvgWorkSessionSeconds))
		fmt.Printf("Average Break: %s\n", formatDurationSeconds(report.AvgBreakSeconds))

		// Display daily breakdown if available
		if len(report.DailyStats) > 0 {
			fmt.Printf("\nDaily Breakdown:\n")
			fmt.Printf("%-12s %-12s %-12s %-12s\n", "Date", "Sessions", "Work Time", "Completion")
			fmt.Printf("%-12s %-12s %-12s %-12s\n", "----", "--------", "---------", "----------")

			for _, day := range report.DailyStats {
				completionRate := 0.0
				if day.TotalSessions > 0 {
					completionRate = float64(day.CompletedSessions) / float64(day.TotalSessions) * 100
				}

				fmt.Printf("%-12s %-12d %-12s %.1f%%\n",
					day.Date.Format("2006-01-02"),
					day.TotalSessions,
					formatDurationSeconds(day.WorkTimeSeconds),
					completionRate)
			}
		}

		// If task filtering is not applied, show top tasks
		if taskID == nil && len(report.TopTasks) > 0 {
			fmt.Printf("\nMost Popular Tasks:\n")
			fmt.Printf("%-5s %-30s %-12s %-12s\n", "ID", "Description", "Sessions", "Total Time")
			fmt.Printf("%-5s %-30s %-12s %-12s\n", "--", "-----------", "--------", "----------")

			for _, task := range report.TopTasks {
				description := task.Description
				if len(description) > 27 {
					description = description[:24] + "..."
				}

				fmt.Printf("%-5d %-30s %-12d %s\n",
					task.ID,
					description,
					task.SessionCount,
					formatDurationSeconds(task.TotalTimeSeconds))
			}
		}

		// If generating a file output, save the report
		if reportOutput == "file" {
			filename := fmt.Sprintf("pomodoro_report_%s.txt", time.Now().Format("2006-01-02"))
			fmt.Printf("\nSaving report to %s\n", filename)
			// Code to save report to file would go here
		}
	},
}

func init() {
	pomoCmd.AddCommand(reportCmd)

	// Add flags
	reportCmd.Flags().StringVar(&reportPeriod, "period", "", "Report period (day, week, month, year)")
	reportCmd.Flags().StringVar(&reportFormat, "format", "text", "Report format (text, json)")
	reportCmd.Flags().StringVar(&reportOutput, "output", "console", "Report output (console, file)")
}
