package cmd

type ActionType string

// Action constants
const (
	DELETE   ActionType = "delete"
	COMPLETE ActionType = "complete"
	START    ActionType = "start"
	PAUSE    ActionType = "pause"
	EDIT     ActionType = "edit"
	DUE      ActionType = "set start date to today for"
	// Add more actions as needed
)
