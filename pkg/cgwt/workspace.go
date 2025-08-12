package cgwt

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// WorkspaceConfig represents a VS Code/Cursor workspace configuration.
type WorkspaceConfig struct {
	Name       string                 `json:"name,omitempty"`
	Folders    []WorkspaceFolder      `json:"folders"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// WorkspaceFolder represents a folder in a workspace configuration.
type WorkspaceFolder struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
}

// detectWorkspaceMode checks if the current directory contains workspace files.
func (c *realCGWT) detectWorkspaceMode() ([]string, error) {
	c.verbosePrint("Checking for .code-workspace files...")

	// Check for workspace files
	workspaceFiles, err := c.fs.Glob("*.code-workspace")
	if err != nil {
		return nil, fmt.Errorf("failed to check for workspace files: %w", err)
	}

	if len(workspaceFiles) == 0 {
		c.verbosePrint("No .code-workspace files found")
		return nil, nil
	}

	c.verbosePrint(fmt.Sprintf("Found %d workspace file(s)", len(workspaceFiles)))
	return workspaceFiles, nil
}

// parseWorkspaceFile parses a workspace configuration file.
func (c *realCGWT) parseWorkspaceFile(filename string) (*WorkspaceConfig, error) {
	c.verbosePrint("Parsing workspace configuration...")

	// Read workspace file
	content, err := c.fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWorkspaceFileReadError, err)
	}

	// Parse JSON
	var config WorkspaceConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, ErrWorkspaceFileMalformed
	}

	// Validate folders array
	if config.Folders == nil {
		return nil, ErrWorkspaceEmptyFolders
	}

	// Filter out null values and validate structure
	var validFolders []WorkspaceFolder
	for _, folder := range config.Folders {
		// Skip null values
		if folder.Path == "" {
			continue
		}

		// Validate path field
		if folder.Path == "" {
			return nil, fmt.Errorf("workspace folder must contain path field")
		}

		// Validate name field if present
		// Name field is optional, but if present it should be a string
		// This is already handled by JSON unmarshaling

		validFolders = append(validFolders, folder)
	}

	// Check if we have any valid folders after filtering
	if len(validFolders) == 0 {
		return nil, ErrWorkspaceEmptyFolders
	}

	config.Folders = validFolders
	return &config, nil
}

// getWorkspaceName extracts the workspace name from configuration or filename.
func (c *realCGWT) getWorkspaceName(config *WorkspaceConfig, filename string) string {
	// First try to get name from workspace configuration
	if config.Name != "" {
		return config.Name
	}

	// Fallback to filename without extension
	return strings.TrimSuffix(filepath.Base(filename), ".code-workspace")
}

// handleMultipleWorkspaces handles the selection of workspace files when multiple are found.
func (c *realCGWT) handleMultipleWorkspaces(workspaceFiles []string) (string, error) {
	c.verbosePrint(fmt.Sprintf("Multiple workspace files found: %d", len(workspaceFiles)))

	// Display selection prompt
	c.displayWorkspaceSelection(workspaceFiles)

	// Get user selection
	selection, err := c.getUserSelection(len(workspaceFiles))
	if err != nil {
		return "", fmt.Errorf("user cancelled selection: %w", err)
	}

	selectedFile := workspaceFiles[selection-1] // Convert to 0-based index

	// Confirm selection
	confirmed, err := c.confirmSelection(selectedFile)
	if err != nil {
		return "", fmt.Errorf("user cancelled confirmation: %w", err)
	}

	if !confirmed {
		// User wants to go back to selection
		return c.handleMultipleWorkspaces(workspaceFiles)
	}

	c.verbosePrint(fmt.Sprintf("Selected workspace file: %s", selectedFile))
	return selectedFile, nil
}

// displayWorkspaceSelection displays the workspace selection prompt.
func (c *realCGWT) displayWorkspaceSelection(workspaceFiles []string) {
	fmt.Println("Multiple workspace files found. Please select one:")
	fmt.Println()

	for i, file := range workspaceFiles {
		fmt.Printf("%d. %s\n", i+1, file)
	}

	fmt.Println()
	fmt.Printf("Enter your choice (1-%d) or 'q' to quit: ", len(workspaceFiles))
}

// getUserSelection gets and validates user input for workspace selection.
func (c *realCGWT) getUserSelection(maxChoice int) (int, error) {
	return c.getUserSelectionWithRetries(maxChoice, 3)
}

// getUserSelectionWithRetries gets and validates user input with retry limit.
func (c *realCGWT) getUserSelectionWithRetries(maxChoice int, retries int) (int, error) {
	if retries <= 0 {
		return 0, fmt.Errorf("too many invalid inputs, user cancelled selection")
	}

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return 0, fmt.Errorf("failed to read user input: %w", err)
	}

	// Handle quit commands
	if c.isQuitCommand(input) {
		return 0, fmt.Errorf("user cancelled selection")
	}

	// Parse numeric input
	choice, err := c.parseNumericInput(input)
	if err != nil {
		fmt.Printf("Please enter a number between 1 and %d or 'q' to quit: ", maxChoice)
		return c.getUserSelectionWithRetries(maxChoice, retries-1)
	}

	// Validate range
	if !c.isValidChoice(choice, maxChoice) {
		fmt.Printf("Please enter a number between 1 and %d or 'q' to quit: ", maxChoice)
		return c.getUserSelectionWithRetries(maxChoice, retries-1)
	}

	return choice, nil
}

// isQuitCommand checks if the input is a quit command.
func (c *realCGWT) isQuitCommand(input string) bool {
	quitCommands := []string{"q", "quit", "exit", "cancel"}
	for _, cmd := range quitCommands {
		if input == cmd {
			return true
		}
	}
	return false
}

// parseNumericInput parses numeric input from string.
func (c *realCGWT) parseNumericInput(input string) (int, error) {
	var choice int
	_, err := fmt.Sscanf(input, "%d", &choice)
	return choice, err
}

// isValidChoice checks if the choice is within valid range.
func (c *realCGWT) isValidChoice(choice, maxChoice int) bool {
	return choice >= 1 && choice <= maxChoice
}

// confirmSelection asks the user to confirm their workspace selection.
func (c *realCGWT) confirmSelection(workspaceFile string) (bool, error) {
	return c.confirmSelectionWithRetries(workspaceFile, 3)
}

// confirmSelectionWithRetries asks the user to confirm their workspace selection with retry limit.
func (c *realCGWT) confirmSelectionWithRetries(workspaceFile string, retries int) (bool, error) {
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
	result, err := c.parseConfirmationInput(input)
	if err != nil {
		fmt.Print("Please enter 'y' for yes, 'n' for no, or 'q' to quit: ")
		return c.confirmSelectionWithRetries(workspaceFile, retries-1)
	}

	return result, nil
}

// parseConfirmationInput parses confirmation input.
func (c *realCGWT) parseConfirmationInput(input string) (bool, error) {
	switch input {
	case "y", "yes", "Y", "YES":
		return true, nil
	case "n", "no", "N", "NO":
		return false, nil
	case "q", "quit", "exit", "cancel":
		return false, fmt.Errorf("user cancelled confirmation")
	default:
		return false, fmt.Errorf("invalid input")
	}
}
