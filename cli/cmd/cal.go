/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// calCmd represents the cal command
var calCmd = &cobra.Command{
	Use:   "cal",
	Short: "Manage your calendar and scheduled events",
	Long: `Create, view, and manage calendar events and appointments.

Calendar events can be linked to projects and tasks to help with scheduling and time management.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("cal called")
	},
}

func init() {
	rootCmd.AddCommand(calCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// calCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// calCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
