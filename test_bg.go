package main

import (
	"fmt"
)

func main() {
	// ANSI color codes
	const (
		reset      = "\033[0m"
		darkGrayBg = "\033[48;5;236m"
	)

	// Print with background color first
	fmt.Println("Without background:")
	fmt.Println("Line 1")
	fmt.Println("Line 2")

	fmt.Println("\nWith background (alternating):")
	fmt.Print(darkGrayBg + "Line 1" + "\033[K" + reset + "\n")
	fmt.Println("Line 2")
	fmt.Print(darkGrayBg + "Line 3" + "\033[K" + reset + "\n")
	fmt.Println("Line 4")

	fmt.Println("\nTesting continuation lines:")
	// Row 1 with continuation
	fmt.Println("Row 1 main line")
	fmt.Println("Row 1 continuation line")

	// Row 2 with continuation (with background)
	fmt.Print(darkGrayBg + "Row 2 main line" + "\033[K" + reset + "\n")
	fmt.Print(darkGrayBg + "Row 2 continuation line" + "\033[K" + reset + "\n")

	// Row 3 without continuation
	fmt.Println("Row 3 (no continuation)")

	// Row 4 with continuation (with background)
	fmt.Print(darkGrayBg + "Row 4 main line" + "\033[K" + reset + "\n")
	fmt.Print(darkGrayBg + "Row 4 continuation line" + "\033[K" + reset + "\n")
}
