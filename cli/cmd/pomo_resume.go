package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume a paused Pomodoro session",
	Long: `Resume a previously paused Pomodoro session.

Examples:
  prod pomo resume  # Resume the currently paused Pomodoro session`,

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
			fmt.Println("You need to be logged in to resume a Pomodoro session")
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

		// Check if session is paused
		if activeSession.Status != services.StatusPaused {
			fmt.Println("Your Pomodoro session is not paused")
			fmt.Println("Use 'prod pomo pause' to pause it first")
			return
		}

		// Resume the session
		resumedSession, err := pomoService.ResumeSession(context.Background(), user.ID)
		if err != nil {
			fmt.Printf("Error resuming Pomodoro session: %v\n", err)
			return
		}

		// Calculate new end time
		startTime := resumedSession.StartTime.Time
		workDuration := resumedSession.WorkDuration
		pauseDuration := resumedSession.TotalPauseDuration
		endTime := startTime.Add(workDuration).Add(pauseDuration)

		// Print result
		fmt.Println("▶️  Pomodoro session resumed")
		fmt.Printf("Started at: %s\n", startTime.Format("15:04:05"))
		fmt.Printf("Current time: %s\n", time.Now().Format("15:04:05"))
		fmt.Printf("Work until: %s\n", endTime.Format("15:04:05"))
		fmt.Printf("Total pause time: %s\n", formatDuration(pauseDuration))

		if resumedSession.TaskID != nil {
			taskService := services.NewTaskService(queries)
			task, _ := taskService.GetTask(context.Background(), *resumedSession.TaskID, user.ID)
			fmt.Printf("Task: %s (ID: %d)\n", task.Description, task.ID)
		}

		fmt.Println("\nUse 'prod pomo stop' when you're done")
	},
}

func init() {
	pomoCmd.AddCommand(resumeCmd)
}
