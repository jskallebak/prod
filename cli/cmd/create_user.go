package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jskallebak/prod/internal/db/sqlc"
	"github.com/jskallebak/prod/internal/services"
	"github.com/jskallebak/prod/internal/util"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	userName string
)

// createUserCmd represents the create user command
var createUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	Long: `Create a new user account in the productivity system.
	
For example:
  prod user create --name "John Doe"
  
You will be prompted to enter an email and password.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize DB connection
		dbpool, err := util.InitDB()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
			os.Exit(1)
		}
		defer dbpool.Close()

		// Get email from user
		fmt.Print("Enter email: ")
		var email string
		fmt.Scanln(&email)
		email = strings.TrimSpace(email)

		// Validate email
		if email == "" {
			fmt.Println("Email cannot be empty")
			return
		}

		// Get password (without echoing to terminal)
		fmt.Print("Enter password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError reading password: %v\n", err)
			return
		}
		password := string(passwordBytes)
		fmt.Println() // Add a newline after password input

		// Confirm password
		fmt.Print("Confirm password: ")
		confirmPasswordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError reading password confirmation: %v\n", err)
			return
		}
		confirmPassword := string(confirmPasswordBytes)
		fmt.Println() // Add a newline after password confirmation input

		// Check password match
		if password != confirmPassword {
			fmt.Println("Passwords do not match")
			return
		}

		// Create queries and service
		queries := sqlc.New(dbpool)
		userService := services.NewUserService(queries)

		// Set up parameters
		params := services.CreateUserParams{
			Email:    email,
			Password: password,
		}

		// Add name if provided
		if cmd.Flags().Changed("name") {
			params.Name = &userName
		}

		// Create the user
		user, err := userService.CreateUser(context.Background(), params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating user: %v\n", err)
			return
		}

		fmt.Printf("User created successfully!\n")
		fmt.Printf("User ID: %d\n", user.ID)
		fmt.Printf("Email: %s\n", user.Email)
		if user.Name.Valid {
			fmt.Printf("Name: %s\n", user.Name.String)
		}
	},
}

func init() {
	userCmd.AddCommand(createUserCmd)

	// Define flags for the create user command
	createUserCmd.Flags().StringVarP(&userName, "name", "n", "", "User's full name")
}
