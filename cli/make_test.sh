#!/bin/zsh

# Script to create a set of tasks in the productivity app
# Including a main task and two subtasks

echo "Creating test tasks..."

# Specify the directory where your Go app is located
# Change this to the actual path of your Go app
APP_DIR="$HOME/workspace/prod/cli"  # Update this path!

# Function to run the command in the correct directory
prod() {
    (cd "$APP_DIR" && go run . "$@")
}

# Create main task
echo "Adding main task..."
prod task add "test"

# Create first subtask
echo "Adding first subtask..."
prod task add "first" -s 4

# Create second subtask
echo "Adding second subtask..."
prod task add "second" -s 4

echo "All tasks created successfully!"
