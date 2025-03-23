/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"syscall"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

var (
	newPassword string
)

// passwordCmd represents the password command
var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize DB and services
		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		authService := services.NewAuthService(queries)

		// Get current user
		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Authentication required: %v\n", err)
			fmt.Println("Please login first with: prod login")
			return
		}

		// Verify current password first
		fmt.Print("Enter current password: ")
		currentPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError reading password: %v\n", err)
			return
		}
		fmt.Println() // Add a newline after password input

		currentPassword := string(currentPasswordBytes)

		// Verify current password matches
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Current password is incorrect\n")
			return
		}

		// Get new password
		fmt.Print("Enter new password: ")
		newPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError reading password: %v\n", err)
			return
		}
		fmt.Println() // Add a newline after password input

		// Get confirmation of new password
		fmt.Print("Confirm new password: ")
		confirmPasswordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError reading password: %v\n", err)
			return
		}
		fmt.Println() // Add a newline after password input

		// Check if passwords match
		if string(newPasswordBytes) != string(confirmPasswordBytes) {
			fmt.Fprintf(os.Stderr, "Passwords do not match\n")
			return
		}

		// Update password
		_, err = authService.UpdatePassword(context.Background(), user.ID, string(newPasswordBytes))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to update password: %v\n", err)
			return
		}

		fmt.Println("Password successfully updated")
	},
}

func init() {
	userCmd.AddCommand(passwordCmd)

	passwordCmd.Flags().StringVarP(&newPassword, "password", "p", "", "New password")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// passwordCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// passwordCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
