/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage user accounts",
	Long: `Manage user accounts in the productivity system.

This command allows you to create and manage user accounts, including:
- Creating new users
- Updating user information
- Changing passwords
- Managing authentication`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use one of the user subcommands. Try 'prod user --help' for more information.")
	},
}

func init() {
	rootCmd.AddCommand(userCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// userCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// userCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
