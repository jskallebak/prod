package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var forceComplete bool

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the current Pomodoro session",
	Long: `Stop the current Pomodoro session and save completion information.

Examples:
  prod pomo stop            # Stop and mark as cancelled
  prod pomo stop --complete # Stop and mark as completed`,

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
			fmt.Println("You need to be logged in to stop a Pomodoro session")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Get the active session
		pomoService := services.NewPomodoroService(queries)
		activeSession, err := pomoService.GetActiveSession(context.Background(), user.ID)
		if err != nil {
			fmt.Println("You don't have an active Pomodoro session")
			return
		}

		// Determine whether to mark as completed
		// If work duration has elapsed, we mark as completed
		complete := forceComplete
		if !complete {
			startTime := activeSession.StartTime.Time
			elapsed := time.Since(startTime)

			// Check if we've met at least 80% of the work duration
			targetDuration := activeSession.WorkDuration
			minDuration := time.Duration(float64(targetDuration) * 0.8)

			// Factor in pause time if the session was paused
			if activeSession.Status == services.StatusPaused {
				// If paused, we subtract pause time from elapsed time
				if activeSession.PauseTime.Valid {
					pauseStart := activeSession.PauseTime.Time
					pauseDuration := time.Since(pauseStart)
					elapsed = elapsed - pauseDuration
				}
			}

			// Also factor in total pause duration
			elapsed = elapsed - activeSession.TotalPauseDuration

			complete = elapsed >= minDuration
		}

		// Stop the session
		stoppedSession, err := pomoService.StopSession(context.Background(), user.ID, complete)
		if err != nil {
			fmt.Printf("Error stopping Pomodoro session: %v\n", err)
			return
		}

		// Print result
		status := "cancelled"
		if complete {
			status = "completed"
		}

		fmt.Printf("🍅 Pomodoro session %s!\n", status)

		// Print session details
		startTime := stoppedSession.StartTime.Time
		endTime := stoppedSession.EndTime.Time
		duration := endTime.Sub(startTime) - stoppedSession.TotalPauseDuration

		fmt.Printf("Started: %s\n", startTime.Format("15:04:05"))
		fmt.Printf("Ended: %s\n", endTime.Format("15:04:05"))
		fmt.Printf("Duration: %s\n", util.FormatDuration(duration))

		if stoppedSession.TaskID != nil {
			taskService := services.NewTaskService(queries)
			task, _ := taskService.GetTask(context.Background(), *stoppedSession.TaskID, user.ID)
			fmt.Printf("Task: %s (ID: %d)\n", task.Description, task.ID)
		}

		if stoppedSession.Note != "" {
			fmt.Printf("Note: %s\n", stoppedSession.Note)
		}
	},
}

func init() {
	pomoCmd.AddCommand(stopCmd)

	// Add flags
	stopCmd.Flags().BoolVar(&forceComplete, "complete", false, "Mark the session as completed regardless of duration")
}
