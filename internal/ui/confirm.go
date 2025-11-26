package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm prompts the user with a yes/no question and returns their response
// Returns true if user responds with 'y' or 'yes' (case insensitive)
// Returns false for 'n', 'no', or any other input
func Confirm(message string) bool {
	fmt.Printf("%s [y/N]: ", message)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// ConfirmWithDefault prompts the user with a yes/no question with a specified default
// If user just presses Enter, returns the default value
func ConfirmWithDefault(message string, defaultYes bool) bool {
	var prompt string
	if defaultYes {
		prompt = fmt.Sprintf("%s [Y/n]: ", message)
	} else {
		prompt = fmt.Sprintf("%s [y/N]: ", message)
	}

	fmt.Print(prompt)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}

	response = strings.ToLower(strings.TrimSpace(response))

	// If empty response, return default
	if response == "" {
		return defaultYes
	}

	return response == "y" || response == "yes"
}
