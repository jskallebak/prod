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
	listLimit  int
	listStatus string
	listDate   string
)

var pomoListCmd = &cobra.Command{
	Use:   "list [task-id]",
	Short: "List Pomodoro sessions",
	Long: `List your Pomodoro sessions, optionally filtered by task ID or date.

Examples:
  prod pomo list            # List recent Pomodoro sessions
  prod pomo list 5          # List Pomodoro sessions for task with ID 5
  prod pomo list --today    # List today's Pomodoro sessions
  prod pomo list --date 2024-06-01  # List sessions for a specific date
  prod pomo list --status completed  # List only completed sessions`,

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
			fmt.Println("You need to be logged in to view Pomodoro sessions")
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

			fmt.Printf("Pomodoro Sessions for Task: %s (ID: %d)\n\n", task.Description, task.ID)
		} else {
			fmt.Println("Recent Pomodoro Sessions")
		}

		// Parse date filter
		var startDate, endDate *time.Time
		today, _ := cmd.Flags().GetBool("today")

		if today {
			now := time.Now()
			start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			startDate = &start
			fmt.Printf("Date: Today (%s)\n", start.Format("2006-01-02"))
		} else if listDate != "" {
			date, err := util.ParseDate(listDate)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid date format: %v. Use YYYY-MM-DD\n", err)
				return
			}

			start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
			end := start.Add(24 * time.Hour)
			startDate = &start
			endDate = &end

			fmt.Printf("Date: %s\n", start.Format("2006-01-02"))
		}

		if listStatus != "" {
			fmt.Printf("Status: %s\n", listStatus)
		}

		fmt.Println()

		// Get Pomodoro sessions
		pomoService := services.NewPomodoroService(queries)
		sessions, err := pomoService.ListSessions(context.Background(), user.ID, taskID, startDate, endDate, listStatus, int32(listLimit))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving Pomodoro sessions: %v\n", err)
			return
		}

		if len(sessions) == 0 {
			fmt.Println("No Pomodoro sessions found matching the criteria")
			return
		}

		// Display sessions
		fmt.Println("ID\tSTART TIME\t\tDURATION\tSTATUS\t\tTASK")
		fmt.Println("--\t----------\t\t--------\t------\t\t----")

		for _, session := range sessions {
			// Format duration
			duration := fmt.Sprintf("%d min", session.WorkDuration.Minutes())

			// Format status
			status := string(session.Status)
			if session.Status == services.StatusActive {
				status = "In Progress"
			} else if session.Status == services.StatusCompleted {
				status = "Completed"
			} else if session.Status == services.StatusCancelled {
				status = "Cancelled"
			} else if session.Status == services.StatusPaused {
				status = "Paused"
			}

			// Format task description
			taskDesc := "-"
			if session.TaskID != nil {
				taskService := services.NewTaskService(queries)
				task, err := taskService.GetTask(context.Background(), *session.TaskID, user.ID)
				if err == nil && task != nil {
					if len(task.Description) > 25 {
						taskDesc = task.Description[:22] + "..."
					} else {
						taskDesc = task.Description
					}
				}
			}

			fmt.Printf("%d\t%s\t%s\t\t%-10s\t%s\n",
				session.ID,
				session.StartTime.Time.Format("2006-01-02 15:04"),
				duration,
				status,
				taskDesc)
		}

		if len(sessions) == listLimit {
			fmt.Printf("\nShowing %d sessions. Use --limit to see more.\n", listLimit)
		}
	},
}

func init() {
	pomoCmd.AddCommand(pomoListCmd)

	// Add flags
	pomoListCmd.Flags().IntVar(&listLimit, "limit", 10, "Maximum number of sessions to show")
	pomoListCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (in_progress, completed, cancelled, paused)")
	pomoListCmd.Flags().StringVar(&listDate, "date", "", "Filter by date (YYYY-MM-DD)")
	pomoListCmd.Flags().Bool("today", false, "Show only today's sessions")
}
