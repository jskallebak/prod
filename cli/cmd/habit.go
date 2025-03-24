/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// habitCmd represents the habit command
var habitCmd = &cobra.Command{
	Use:   "habit",
	Short: "Track and manage your habits and routines",
	Long: `Create, track, and analyze habits and regular routines.

Set up habit tracking with customizable schedules, streaks, and reporting to build consistent behaviors.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("habit called")
	},
}

func init() {
	rootCmd.AddCommand(habitCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// habitCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// habitCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
