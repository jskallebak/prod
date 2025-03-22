/*
Copyright © 2025 Jakob Skallebak jskallebak@gmail.com
*/
package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/jskallebak/prod/cli/cmd"
)

func main() {
	fmt.Println("Program started")

	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	cmd.Execute()
}
