package services

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jskallebak/prod/internal/db/sqlc"
)

// RecurrenceType defines frequency types for recurring tasks
type RecurrenceType string

const (
	RecurrenceDaily   RecurrenceType = "daily"
	RecurrenceWeekly  RecurrenceType = "weekly"
	RecurrenceMonthly RecurrenceType = "monthly"
	RecurrenceYearly  RecurrenceType = "yearly"
)

// RecurrencePattern represents a structured version of the recurrence rule
type RecurrencePattern struct {
	Type         RecurrenceType // daily, weekly, monthly, yearly
	Interval     int            // every X days/weeks/months/years
	WeekDays     []int          // for weekly: days of week (1-7, where 1=Monday)
	MonthDay     int            // for monthly: day of month (1-31 or -1 for last day)
	MonthWeekDay *struct {      // for monthly by weekday
		Week    int // 1-5 (5 means last)
		WeekDay int // 1-7 (1=Monday)
	}
	YearlyDate *struct { // for yearly
		Month int // 1-12
		Day   int // 1-31
	}
	Until *time.Time // recur until this date (optional)
	Count int        // recur this many times (optional)
}

// RecurrenceState tracks the state of a recurring task
type RecurrenceState struct {
	Pattern     RecurrencePattern
	OriginalDue time.Time // The original due date that sets the pattern
	LastCreated time.Time // Last instance that was created
	InstanceNum int       // How many instances have been created so far
}

// ParseRecurrence converts a recurrence string to a structured RecurrencePattern
// ParseRecurrence converts a recurrence string to a structured RecurrencePattern
func ParseRecurrence(recurrence string) (*RecurrencePattern, error) {
	parts := strings.Split(recurrence, ":")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid recurrence format")
	}

	// Start with default values
	pattern := &RecurrencePattern{
		Interval: 1,
		Count:    0, // 0 means no count limit
	}

	// Parse frequency
	frequency := strings.ToLower(parts[0])
	switch frequency {
	case "daily", "weekly", "monthly", "yearly":
		pattern.Type = RecurrenceType(frequency)
	default:
		return nil, fmt.Errorf("unknown frequency: %s", frequency)
	}

	// Parse interval if provided
	if len(parts) > 1 && parts[1] != "" {
		interval, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid interval: %s", parts[1])
		}
		if interval < 1 {
			return nil, fmt.Errorf("interval must be at least 1")
		}
		pattern.Interval = interval
	}

	// Parse frequency-specific details
	if len(parts) > 2 && parts[2] != "" {
		switch pattern.Type {
		case RecurrenceWeekly:
			// Parse weekdays (e.g., "1,3,5" for Monday, Wednesday, Friday)
			dayStrings := strings.Split(parts[2], ",")
			for _, dayStr := range dayStrings {
				day, err := strconv.Atoi(dayStr)
				if err != nil {
					return nil, fmt.Errorf("invalid weekday: %s", dayStr)
				}
				if day < 1 || day > 7 {
					return nil, fmt.Errorf("weekday must be between 1 and 7")
				}
				pattern.WeekDays = append(pattern.WeekDays, day)
			}

		case RecurrenceMonthly:
			// Parse monthly options
			detail := parts[2]

			// Check if it's a week pattern like "2w3" (2nd Wednesday)
			if strings.Contains(detail, "w") {
				wParts := strings.Split(detail, "w")
				if len(wParts) != 2 {
					return nil, fmt.Errorf("invalid month week pattern: %s", detail)
				}

				week, err := strconv.Atoi(wParts[0])
				if err != nil || week < 1 || week > 5 {
					return nil, fmt.Errorf("week must be between 1 and 5")
				}

				weekDay, err := strconv.Atoi(wParts[1])
				if err != nil || weekDay < 1 || weekDay > 7 {
					return nil, fmt.Errorf("weekday must be between 1 and 7")
				}

				pattern.MonthWeekDay = &struct {
					Week    int
					WeekDay int
				}{
					Week:    week,
					WeekDay: weekDay,
				}
			} else if detail == "last" {
				// Last day of month
				pattern.MonthDay = -1
			} else {
				// Specific day of month
				day, err := strconv.Atoi(detail)
				if err != nil {
					return nil, fmt.Errorf("invalid day of month: %s", detail)
				}
				if day < 1 || day > 31 {
					return nil, fmt.Errorf("day must be between 1 and 31")
				}
				pattern.MonthDay = day
			}

		case RecurrenceYearly:
			// Parse yearly date (MMDD format)
			if len(parts[2]) != 4 {
				return nil, fmt.Errorf("yearly date must be in MMDD format")
			}

			monthStr := parts[2][:2]
			dayStr := parts[2][2:]

			month, err := strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				return nil, fmt.Errorf("month must be between 01 and 12")
			}

			day, err := strconv.Atoi(dayStr)
			if err != nil || day < 1 || day > 31 {
				return nil, fmt.Errorf("day must be between 01 and 31")
			}

			pattern.YearlyDate = &struct {
				Month int
				Day   int
			}{
				Month: month,
				Day:   day,
			}
		}
	}

	// Process additional parts for until/count
	for i := 3; i < len(parts); i += 2 {
		if i+1 >= len(parts) {
			return nil, fmt.Errorf("incomplete until/count specification")
		}

		specifier := strings.ToLower(parts[i])
		value := parts[i+1]

		switch specifier {
		case "until":
			// Parse until date (YYYY-MM-DD)
			untilDate, err := time.Parse("2006-01-02", value)
			if err != nil {
				return nil, fmt.Errorf("invalid until date format: %s", value)
			}
			pattern.Until = &untilDate

		case "count":
			// Parse count limit
			count, err := strconv.Atoi(value)
			if err != nil || count < 1 {
				return nil, fmt.Errorf("count must be a positive integer")
			}
			pattern.Count = count

		default:
			return nil, fmt.Errorf("unknown specifier: %s", specifier)
		}
	}

	return pattern, nil
}

// GetNextOccurrence calculates the next occurrence of a recurring task
// GetNextOccurrence calculates the next occurrence of a recurring task
func GetNextOccurrence(pattern RecurrencePattern, referenceDate time.Time) (time.Time, error) {
	var nextDate time.Time

	switch pattern.Type {
	case RecurrenceDaily:
		// For daily recurrence, simply add the interval in days
		nextDate = referenceDate.AddDate(0, 0, pattern.Interval)

	case RecurrenceWeekly:
		// For weekly recurrence, handle day-of-week specifications
		if len(pattern.WeekDays) == 0 {
			// If no specific days are specified, use the same day of the week
			nextDate = referenceDate.AddDate(0, 0, 7*pattern.Interval)
		} else {
			// Get the current day of week (1=Monday in our format)
			currentDayOfWeek := int(referenceDate.Weekday())
			if currentDayOfWeek == 0 { // Sunday in Go is 0
				currentDayOfWeek = 7 // Convert to our format where Sunday is 7
			}

			// Find the next day in the specified list
			nextDayFound := false
			daysToAdd := 0

			// Sort weekdays to ensure proper sequence
			sortedWeekDays := make([]int, len(pattern.WeekDays))
			copy(sortedWeekDays, pattern.WeekDays)
			sort.Ints(sortedWeekDays)

			// First check if there's a later day in the current week
			for _, day := range sortedWeekDays {
				if day > currentDayOfWeek {
					daysToAdd = day - currentDayOfWeek
					nextDayFound = true
					break
				}
			}

			// If not, take the first day in the next interval
			if !nextDayFound {
				daysToAdd = sortedWeekDays[0] - currentDayOfWeek + 7*pattern.Interval
			}

			nextDate = referenceDate.AddDate(0, 0, daysToAdd)
		}

	case RecurrenceMonthly:
		// For monthly recurrence, handle day-of-month or week-of-month patterns
		if pattern.MonthDay > 0 {
			// Specific day of month
			currentMonth := referenceDate.Month()
			currentYear := referenceDate.Year()

			// Start with the same day in the next interval month
			nextMonth := int(currentMonth) + pattern.Interval
			nextYear := currentYear

			// Adjust year if needed
			for nextMonth > 12 {
				nextMonth -= 12
				nextYear++
			}

			// Create the date with the specified day
			nextDate = time.Date(nextYear, time.Month(nextMonth), pattern.MonthDay,
				referenceDate.Hour(), referenceDate.Minute(),
				referenceDate.Second(), 0, referenceDate.Location())

			// Adjust if the day doesn't exist in that month
			if nextDate.Day() != pattern.MonthDay {
				// The specified day doesn't exist in this month (e.g., Feb 30)
				// Go back to the last day of the month
				lastDay := time.Date(nextYear, time.Month(nextMonth+1), 0,
					referenceDate.Hour(), referenceDate.Minute(),
					referenceDate.Second(), 0, referenceDate.Location())
				nextDate = lastDay
			}
		} else if pattern.MonthDay == -1 {
			// Last day of month
			currentMonth := referenceDate.Month()
			currentYear := referenceDate.Year()

			// Calculate target month
			nextMonth := int(currentMonth) + pattern.Interval
			nextYear := currentYear

			// Adjust year if needed
			for nextMonth > 12 {
				nextMonth -= 12
				nextYear++
			}

			// Get the last day of the target month
			nextDate = time.Date(nextYear, time.Month(nextMonth+1), 0,
				referenceDate.Hour(), referenceDate.Minute(),
				referenceDate.Second(), 0, referenceDate.Location())
		} else if pattern.MonthWeekDay != nil {
			// Specific weekday in a specific week (e.g., 2nd Wednesday)
			week := pattern.MonthWeekDay.Week
			weekday := pattern.MonthWeekDay.WeekDay

			// Convert to Go's weekday format (0=Sunday)
			goWeekday := weekday % 7 // Convert 7 (Sunday) to 0

			currentMonth := referenceDate.Month()
			currentYear := referenceDate.Year()

			// Calculate target month
			nextMonth := int(currentMonth) + pattern.Interval
			nextYear := currentYear

			// Adjust year if needed
			for nextMonth > 12 {
				nextMonth -= 12
				nextYear++
			}

			// Find the specified weekday in the target month
			if week == 5 {
				// "Last" weekday of the month
				// Start with the last day of the month
				lastDay := time.Date(nextYear, time.Month(nextMonth+1), 0,
					referenceDate.Hour(), referenceDate.Minute(),
					referenceDate.Second(), 0, referenceDate.Location())

				// Move backward until we find the right weekday
				daysToSubtract := (int(lastDay.Weekday()) - goWeekday + 7) % 7
				nextDate = lastDay.AddDate(0, 0, -daysToSubtract)
			} else {
				// Nth weekday of the month
				// Start with the first day of the month
				firstDay := time.Date(nextYear, time.Month(nextMonth), 1,
					referenceDate.Hour(), referenceDate.Minute(),
					referenceDate.Second(), 0, referenceDate.Location())

				// Find the first occurrence of the weekday
				daysToAdd := (goWeekday - int(firstDay.Weekday()) + 7) % 7
				firstOccurrence := firstDay.AddDate(0, 0, daysToAdd)

				// Add the necessary weeks
				nextDate = firstOccurrence.AddDate(0, 0, 7*(week-1))

				// If we've gone into the next month, go back to the last occurrence in the correct month
				if nextDate.Month() != time.Month(nextMonth) {
					nextDate = nextDate.AddDate(0, 0, -7)
				}
			}
		} else {
			return time.Time{}, fmt.Errorf("monthly recurrence requires day or week specification")
		}

	case RecurrenceYearly:
		// For yearly recurrence, handle month and day specifications
		if pattern.YearlyDate != nil {
			currentYear := referenceDate.Year()
			targetYear := currentYear + pattern.Interval

			// Create the date for the specified month and day in the target year
			nextDate = time.Date(targetYear, time.Month(pattern.YearlyDate.Month),
				pattern.YearlyDate.Day, referenceDate.Hour(), referenceDate.Minute(),
				referenceDate.Second(), 0, referenceDate.Location())
		} else {
			// If no specific date is provided, use the same month and day
			nextDate = referenceDate.AddDate(pattern.Interval, 0, 0)
		}

	default:
		return time.Time{}, fmt.Errorf("unsupported recurrence type: %s", pattern.Type)
	}

	// Check if we've reached until date or count limit
	if pattern.Until != nil && nextDate.After(*pattern.Until) {
		return time.Time{}, fmt.Errorf("recurrence has ended (until date reached)")
	}

	// Note: The count limit will need to be checked separately as it requires
	// tracking how many instances have been created so far

	return nextDate, nil
}

// GenerateNextTaskInstance creates the next instance of a recurring task
func GenerateNextTaskInstance(task sqlc.Task, state *RecurrenceState) (*sqlc.Task, error) {
	// If there's no recurrence pattern or it's not valid, return an error
	if !task.Recurrence.Valid || task.Recurrence.String == "" {
		return nil, fmt.Errorf("task does not have a valid recurrence pattern")
	}

	// If this is the first time generating an instance (no state provided)
	if state == nil {
		// Parse the recurrence pattern
		pattern, err := ParseRecurrence(task.Recurrence.String)
		if err != nil {
			return nil, fmt.Errorf("invalid recurrence pattern: %w", err)
		}

		// Use the current task's due date as the reference
		var referenceDate time.Time
		if task.DueDate.Valid {
			referenceDate = task.DueDate.Time
		} else if task.StartDate.Valid {
			referenceDate = task.StartDate.Time
		} else {
			// If neither is set, use current time
			referenceDate = time.Now()
		}

		// Initialize the recurrence state
		state = &RecurrenceState{
			Pattern:     *pattern,
			OriginalDue: referenceDate,
			LastCreated: referenceDate,
			InstanceNum: 1, // The original task is the first instance
		}
	}

	// Check if we've reached the count limit
	if state.Pattern.Count > 0 && state.InstanceNum >= state.Pattern.Count {
		return nil, fmt.Errorf("recurrence has ended (count limit reached)")
	}

	// Calculate the next due date
	nextDue, err := GetNextOccurrence(state.Pattern, state.LastCreated)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate next occurrence: %w", err)
	}

	// Create a new task instance
	newTask := sqlc.Task{
		// Copy user ID from the original task
		UserID: task.UserID,

		// Copy task details
		Description: task.Description,
		Status:      "pending", // New instances always start as pending
		Priority:    task.Priority,
		ProjectID:   task.ProjectID,
		Recurrence:  task.Recurrence,
		Tags:        task.Tags,
		Notes:       task.Notes,
		Dependent:   task.Dependent,

		// Set the new due date
		DueDate: pgtype.Timestamptz{
			Time:  nextDue,
			Valid: true,
		},
	}

	// If the original task had a start date relative to the due date,
	// calculate a relative start date for the new task
	if task.StartDate.Valid && task.DueDate.Valid {
		// Calculate how many days before the due date the start date was
		daysBefore := int(task.DueDate.Time.Sub(task.StartDate.Time).Hours() / 24)

		// Set the new start date to be the same number of days before the new due date
		if daysBefore > 0 {
			newTask.StartDate = pgtype.Timestamptz{
				Time:  nextDue.AddDate(0, 0, -daysBefore),
				Valid: true,
			}
		}
	}

	// Update recurrence state for future reference
	state.LastCreated = nextDue
	state.InstanceNum++

	return &newTask, nil
}
