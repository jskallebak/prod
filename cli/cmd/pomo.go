/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// pomoCmd represents the pomodoro command
var pomoCmd = &cobra.Command{
	Use:   "pomo",
	Short: "Manage Pomodoro sessions",
	Long: `Manage Pomodoro sessions for time management and productivity.

Available Commands:
  start       Start a new Pomodoro session
  stop        Stop the current Pomodoro session
  pause       Pause the current Pomodoro session
  resume      Resume a paused Pomodoro session
  status      Show the status of the current Pomodoro session
  stats       Show Pomodoro statistics
  list        List completed Pomodoro sessions
  report      Generate reports on Pomodoro usage
  config      Configure Pomodoro settings
  attach      Attach a task to the current Pomodoro session
  detach      Remove task attachment from current Pomodoro`,
}

func init() {
	rootCmd.AddCommand(pomoCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pomoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// pomoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
