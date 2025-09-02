// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"fmt"
	"strings"
)

// HandleMultipleFiles handles the selection of workspace files when multiple are found.
func (w *realWorkspace) HandleMultipleFiles(workspaceFiles []string, force bool) (string, error) {
	w.VerbosePrint("Multiple workspace files found: %d", len(workspaceFiles))

	// If force is true, automatically select the first workspace file
	if force {
		selectedFile := workspaceFiles[0]
		w.VerbosePrint("Force mode: automatically selected workspace file: %s", selectedFile)
		return selectedFile, nil
	}

	// Display selection prompt
	w.displaySelection(workspaceFiles)

	// Get user selection
	selection, err := w.getUserSelection(len(workspaceFiles))
	if err != nil {
		return "", fmt.Errorf("user cancelled selection: %w", err)
	}

	selectedFile := workspaceFiles[selection-1] // Convert to 0-based index

	// Confirm selection
	confirmed, err := w.confirmSelection(selectedFile)
	if err != nil {
		return "", fmt.Errorf("user cancelled confirmation: %w", err)
	}

	if !confirmed {
		// User wants to go back to selection
		return w.HandleMultipleFiles(workspaceFiles, force)
	}

	w.VerbosePrint("Selected workspace file: %s", selectedFile)
	return selectedFile, nil
}

// displaySelection displays the workspace selection prompt.
func (w *realWorkspace) displaySelection(workspaceFiles []string) {
	fmt.Println("Multiple workspace files found. Please select one:")
	fmt.Println()

	for i, file := range workspaceFiles {
		fmt.Printf("%d. %s\n", i+1, file)
	}

	fmt.Println()
	fmt.Printf("Enter your choice (1-%d) or 'q' to quit: ", len(workspaceFiles))
}

// getUserSelection gets and validates user input for workspace selection.
func (w *realWorkspace) getUserSelection(maxChoice int) (int, error) {
	return w.getUserSelectionWithRetries(maxChoice, 3)
}

// getUserSelectionWithRetries gets and validates user input with retry limit.
func (w *realWorkspace) getUserSelectionWithRetries(maxChoice int, retries int) (int, error) {
	if retries <= 0 {
		return 0, fmt.Errorf("too many invalid inputs, user cancelled selection")
	}

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return 0, fmt.Errorf("failed to read user input: %w", err)
	}

	// Handle quit commands
	if w.isQuitCommand(input) {
		return 0, fmt.Errorf("user cancelled selection")
	}

	// Parse numeric input
	choice, err := w.parseNumericInput(input)
	if err != nil {
		fmt.Printf("Please enter a number between 1 and %d or 'q' to quit: ", maxChoice)
		return w.getUserSelectionWithRetries(maxChoice, retries-1)
	}

	// Validate range
	if !w.isValidChoice(choice, maxChoice) {
		fmt.Printf("Please enter a number between 1 and %d or 'q' to quit: ", maxChoice)
		return w.getUserSelectionWithRetries(maxChoice, retries-1)
	}

	return choice, nil
}

// confirmSelection asks the user to confirm their workspace selection.
func (w *realWorkspace) confirmSelection(workspaceFile string) (bool, error) {
	return w.confirmSelectionWithRetries(workspaceFile, 3)
}

// confirmSelectionWithRetries asks the user to confirm their workspace selection with retry limit.
func (w *realWorkspace) confirmSelectionWithRetries(workspaceFile string, retries int) (bool, error) {
	if retries <= 0 {
		return false, fmt.Errorf("too many invalid inputs, user cancelled confirmation")
	}

	fmt.Printf("You selected: %s\n", workspaceFile)
	fmt.Println()
	fmt.Print("Proceed with this workspace? (y/n): ")

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return false, fmt.Errorf("failed to read user input: %w", err)
	}

	// Handle confirmation
	result, err := w.parseConfirmationInput(input)
	if err != nil {
		fmt.Print("Please enter 'y' for yes, 'n' for no, or 'q' to quit: ")
		return w.confirmSelectionWithRetries(workspaceFile, retries-1)
	}

	return result, nil
}

// isQuitCommand checks if the input is a quit command.
func (w *realWorkspace) isQuitCommand(input string) bool {
	return strings.ToLower(strings.TrimSpace(input)) == "q" ||
		strings.ToLower(strings.TrimSpace(input)) == "quit" ||
		strings.ToLower(strings.TrimSpace(input)) == "exit"
}

// parseNumericInput parses numeric input from user.
func (w *realWorkspace) parseNumericInput(input string) (int, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, fmt.Errorf("empty input")
	}

	// Check for quit command
	if w.isQuitCommand(input) {
		return -1, nil
	}

	// Parse number
	var choice int
	_, err := fmt.Sscanf(input, "%d", &choice)
	if err != nil {
		return 0, fmt.Errorf("invalid numeric input: %s", input)
	}

	return choice, nil
}

// isValidChoice checks if the choice is valid for the given range.
func (w *realWorkspace) isValidChoice(choice, maxChoice int) bool {
	return choice >= 1 && choice <= maxChoice
}

// parseConfirmationInput parses confirmation input from user.
func (w *realWorkspace) parseConfirmationInput(input string) (bool, error) {
	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid confirmation input: %s", input)
	}
}
