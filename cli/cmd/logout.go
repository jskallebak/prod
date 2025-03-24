/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/jskallebak/prod/internal/auth"
	"github.com/spf13/cobra"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out and remove authentication token",
	Long: `Log out of the productivity app and remove your authentication token.
This will require you to log in again to access your tasks and data.

Example:
  prod logout`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("logging out...")

		err := auth.RemoveToken()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error doing logout: %v\n", err)
			return
		}
		fmt.Println("Successfully logged out.")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// logoutCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// logoutCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
