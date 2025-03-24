package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var (
	configWorkDuration      int
	configBreakDuration     int
	configLongBreakDuration int
	configLongBreakInterval int
	configAutoStartBreaks   bool
	configAutoStartPomos    bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Pomodoro settings",
	Long: `Configure your Pomodoro timer settings.

Examples:
  prod pomo config                         # Show current configuration
  prod pomo config --work 25 --break 5     # Set work and break durations
  prod pomo config --long-break 15         # Set long break duration
  prod pomo config --auto-breaks           # Enable automatic break start`,

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
			fmt.Println("You need to be logged in to configure Pomodoro settings")
			fmt.Println("Use 'prod login' to authenticate")
			return
		}

		// Get Pomodoro service
		pomoService := services.NewPomodoroService(queries)

		// Get current configuration
		currentConfig, err := pomoService.GetUserConfig(context.Background(), user.ID)
		if err != nil {
			// If no config exists, use default values
			currentConfig = &services.PomodoroConfig{
				UserID:             user.ID,
				WorkDuration:       25,
				BreakDuration:      5,
				LongBreakDuration:  15,
				LongBreakInterval:  4,
				AutoStartBreaks:    false,
				AutoStartPomodoros: false,
			}
		}

		// Check if any flag was set
		workFlag := cmd.Flags().Changed("work")
		breakFlag := cmd.Flags().Changed("break")
		longBreakFlag := cmd.Flags().Changed("long-break")
		intervalFlag := cmd.Flags().Changed("interval")
		autoBreaksFlag := cmd.Flags().Changed("auto-breaks")
		autoPomoFlag := cmd.Flags().Changed("auto-pomos")

		// If no flags were provided, just show current settings
		if !workFlag && !breakFlag && !longBreakFlag && !intervalFlag && !autoBreaksFlag && !autoPomoFlag {
			showConfig(currentConfig)
			return
		}

		// Update configuration based on provided flags
		if workFlag {
			if configWorkDuration < 1 {
				fmt.Println("Work duration must be at least 1 minute")
				return
			}
			currentConfig.WorkDuration = int32(configWorkDuration)
		}

		if breakFlag {
			if configBreakDuration < 1 {
				fmt.Println("Break duration must be at least 1 minute")
				return
			}
			currentConfig.BreakDuration = int32(configBreakDuration)
		}

		if longBreakFlag {
			if configLongBreakDuration < 1 {
				fmt.Println("Long break duration must be at least 1 minute")
				return
			}
			currentConfig.LongBreakDuration = int32(configLongBreakDuration)
		}

		if intervalFlag {
			if configLongBreakInterval < 1 {
				fmt.Println("Long break interval must be at least 1")
				return
			}
			currentConfig.LongBreakInterval = int32(configLongBreakInterval)
		}

		if autoBreaksFlag {
			currentConfig.AutoStartBreaks = configAutoStartBreaks
		}

		if autoPomoFlag {
			currentConfig.AutoStartPomodoros = configAutoStartPomos
		}

		// Save the updated configuration
		updatedConfig, err := pomoService.UpdateUserConfig(
			context.Background(),
			user.ID,
			currentConfig.WorkDuration,
			currentConfig.BreakDuration,
			currentConfig.LongBreakDuration,
			currentConfig.LongBreakInterval,
			currentConfig.AutoStartBreaks,
			currentConfig.AutoStartPomodoros,
		)
		if err != nil {
			fmt.Printf("Error updating Pomodoro configuration: %v\n", err)
			return
		}

		fmt.Println("Configuration updated successfully!")
		showConfig(updatedConfig)
	},
}

func showConfig(config *services.PomodoroConfig) {
	fmt.Println("Current Pomodoro Configuration:")
	fmt.Printf("Work Duration:        %d minutes\n", config.WorkDuration)
	fmt.Printf("Break Duration:       %d minutes\n", config.BreakDuration)
	fmt.Printf("Long Break Duration:  %d minutes\n", config.LongBreakDuration)
	fmt.Printf("Long Break Interval:  Every %d pomodoros\n", config.LongBreakInterval)
	fmt.Printf("Auto-start Breaks:    %s\n", strconv.FormatBool(config.AutoStartBreaks))
	fmt.Printf("Auto-start Pomodoros: %s\n", strconv.FormatBool(config.AutoStartPomodoros))
}

func init() {
	pomoCmd.AddCommand(configCmd)

	// Add flags
	configCmd.Flags().IntVar(&configWorkDuration, "work", 0, "Work duration in minutes")
	configCmd.Flags().IntVar(&configBreakDuration, "break", 0, "Break duration in minutes")
	configCmd.Flags().IntVar(&configLongBreakDuration, "long-break", 0, "Long break duration in minutes")
	configCmd.Flags().IntVar(&configLongBreakInterval, "interval", 0, "Number of pomodoros before a long break")
	configCmd.Flags().BoolVar(&configAutoStartBreaks, "auto-breaks", false, "Automatically start breaks after work sessions")
	configCmd.Flags().BoolVar(&configAutoStartPomos, "auto-pomos", false, "Automatically start next pomodoro after breaks")
}
