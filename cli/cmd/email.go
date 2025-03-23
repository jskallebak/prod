/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
)

var (
	newEmail string
)

// emailCmd represents the email command
var emailCmd = &cobra.Command{
	Use:   "email",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("email called")

		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		authService := services.NewAuthService(queries)

		user, err := authService.GetCurrentUser(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Authentication required: %v\n", err)
			fmt.Println("Please login first with: prod login")
			return
		}

		if newEmail == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter email: ")
			email, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading email: %v\n", err)
				return
			}
			newEmail = strings.TrimSpace(email)
		}

		updatedUser, err := authService.UpdateEmail(context.Background(), user.ID, newEmail)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to upddate email: %v\n", err)
			return
		}

		fmt.Printf("Email successfully update to %s\n", updatedUser.Email)

	},
}

func init() {
	userCmd.AddCommand(emailCmd)

	emailCmd.Flags().StringVarP(&newEmail, "email", "e", "", "New email address")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// emailCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// emailCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
