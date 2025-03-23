/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// taskCmd represents the task command
var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks",
	Long:  `Create, list, edit, and manage tasks in the productivity app.`,
	// This runs if 'task' is used with an unknown subcommand
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no args were provided, show help
		if len(args) == 0 {
			return cmd.Help()
		}

		// If args were provided but don't match any subcommand
		availableCommands := []string{}
		for _, subCmd := range cmd.Commands() {
			availableCommands = append(availableCommands, subCmd.Name())
		}

		return fmt.Errorf("unknown command")
	},
}

func init() {
	rootCmd.AddCommand(taskCmd)

	taskCmd.SuggestionsMinimumDistance = 2

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// taskCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// taskCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
