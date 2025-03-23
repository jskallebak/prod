package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	loginEmail    string
	loginPassword string
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the productivity app",
	Long: `Log in to the productivity app to access your tasks, notes, and more.
This command will prompt for your email and password if not provided.`,
	Run: func(cmd *cobra.Command, args []string) {

		dbpool, queries, ok := util.InitDBAndQueriesCLI()
		if !ok {
			return
		}
		defer dbpool.Close()

		auth := services.NewAuthService(queries)

		// For email prompt
		if loginEmail == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter email: ")
			email, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading email: %v\n", err)
				return
			}
			loginEmail = strings.TrimSpace(email)
		}

		// For password prompt
		if loginPassword == "" {
			fmt.Print("Enter password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
				return
			}
			fmt.Println() // Add a newline after password input
			loginPassword = string(passwordBytes)
		}

		_, err := auth.Login(context.Background(), loginEmail, loginPassword)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Login failed: %v\n", err)
			return
		}

		fmt.Println("Login successful! You are now authenticated.")

	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	// Add flags for email and password (optional)
	loginCmd.Flags().StringVarP(&loginEmail, "email", "e", "", "Your email address")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Your password (not recommended for security reasons)")
}
