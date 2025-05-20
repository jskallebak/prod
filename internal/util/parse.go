package util

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseArgs parses arguments in formats like "1", "1-3", "1,2,4", "1 2 4" or combinations
// Returns a slice of integers containing all the specified values
func ParseArgs(args []string) ([]int, error) {
	argStr := strings.Join(args, " ")
	if argStr == "" {
		return nil, fmt.Errorf("empty argument string")
	}

	var result []int

	// First split by commas
	commaSeparated := strings.Split(argStr, ",")

	for _, commaPart := range commaSeparated {
		// Then split each comma-separated part by spaces
		spaceParts := strings.Fields(commaPart)

		for _, part := range spaceParts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			// Check if it's a range (e.g., "1-3")
			if strings.Contains(part, "-") {
				rangeParts := strings.Split(part, "-")
				if len(rangeParts) != 2 {
					return nil, fmt.Errorf("invalid range format: %s", part)
				}

				start, err := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
				if err != nil {
					return nil, fmt.Errorf("invalid range start: %s", rangeParts[0])
				}

				end, err := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
				if err != nil {
					return nil, fmt.Errorf("invalid range end: %s", rangeParts[1])
				}

				if end < start {
					return nil, fmt.Errorf("range end cannot be less than start: %s", part)
				}

				for i := start; i <= end; i++ {
					result = append(result, i)
				}
			} else {
				// Single number
				num, err := strconv.Atoi(part)
				if err != nil {
					return nil, fmt.Errorf("invalid number: %s", part)
				}
				result = append(result, num)
			}
		}
	}

	return result, nil
}

func Input2Int(input string) (int, error) {
	i, err := strconv.Atoi(input)
	if err != nil {
		return 0, err
	}
	return i, nil
}
