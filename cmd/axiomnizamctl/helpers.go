package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// promptInput reads a single line from user input
func promptInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt + ": ")
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// promptPassword reads password without echoing to terminal
func promptPassword(prompt string) string {
	fmt.Print(prompt + ": ")
	bytes, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return string(bytes)
}

// confirmAction asks for yes/no confirmation
func confirmAction(message string) bool {
	fmt.Print(message + " (yes/no): ")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "yes" || response == "y"
}

// printTable prints a simple text table
func printTable(headers []string, rows [][]string) {
	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, header := range headers {
		fmt.Printf("%-*s", colWidths[i]+2, header)
	}
	fmt.Println()

	// Print separator
	for _, width := range colWidths {
		fmt.Print(strings.Repeat("-", width+2))
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			fmt.Printf("%-*s", colWidths[i]+2, cell)
		}
		fmt.Println()
	}
}
