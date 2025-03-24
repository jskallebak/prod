package cmd

import (
	"context"
	"fmt"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var pauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause the current Pomodoro session",
	Long: `Pause the current active Pomodoro session.

Examples:
  prod pomo pause  # Pause the current Pomodoro session`,

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
			fmt.Println("You need to be logged in to pause a Pomodoro session")
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

		// Check if session is already paused
		if activeSession.Status == services.StatusPaused {
			fmt.Println("Pomodoro session is already paused")
			fmt.Println("Use 'prod pomo resume' to resume it")
			return
		}

		// Pause the session
		pausedSession, err := pomoService.PauseSession(context.Background(), user.ID)
		if err != nil {
			fmt.Printf("Error pausing Pomodoro session: %v\n", err)
			return
		}

		// Print result
		fmt.Println("⏸️  Pomodoro session paused")
		fmt.Printf("Started at: %s\n", pausedSession.StartTime.Time.Format("15:04:05"))
		fmt.Printf("Paused at: %s\n", pausedSession.PauseTime.Time.Format("15:04:05"))

		if pausedSession.TaskID != nil {
			taskService := services.NewTaskService(queries)
			task, _ := taskService.GetTask(context.Background(), *pausedSession.TaskID, user.ID)
			fmt.Printf("Task: %s (ID: %d)\n", task.Description, task.ID)
		}

		fmt.Println("\nUse 'prod pomo resume' to continue your session")
		fmt.Println("Use 'prod pomo stop' to end your session")
	},
}

func init() {
	pomoCmd.AddCommand(pauseCmd)
}
