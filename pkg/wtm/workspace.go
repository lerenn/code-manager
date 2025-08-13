package wtm

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/logger"
)

// workspaceConfig represents a VS Code-like workspace configuration.
type workspaceConfig struct {
	Name       string                 `json:"name,omitempty"`
	Folders    []workspaceFolder      `json:"folders"`
	Settings   map[string]interface{} `json:"settings,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// workspaceFolder represents a folder in a workspace configuration.
type workspaceFolder struct {
	Name string `json:"name,omitempty"`
	Path string `json:"path"`
}

// workspace represents a workspace and provides methods for workspace operations.
type workspace struct {
	fs           fs.FS
	git          git.Git
	logger       logger.Logger
	verbose      bool
	originalFile string
}

// newWorkspace creates a new Workspace instance.
func newWorkspace(fs fs.FS, git git.Git, logger logger.Logger, verbose bool) *workspace {
	return &workspace{
		fs:      fs,
		git:     git,
		logger:  logger,
		verbose: verbose,
	}
}

// detectWorkspaceFiles checks if the current directory contains workspace files.
func (w *workspace) detectWorkspaceFiles() ([]string, error) {
	w.verbosePrint("Checking for .code-workspace files...")

	// Check for workspace files
	workspaceFiles, err := w.fs.Glob("*.code-workspace")
	if err != nil {
		return nil, fmt.Errorf("failed to check for workspace files: %w", err)
	}

	if len(workspaceFiles) == 0 {
		w.verbosePrint("No .code-workspace files found")
		return nil, nil
	}

	w.verbosePrint(fmt.Sprintf("Found %d workspace file(s)", len(workspaceFiles)))
	return workspaceFiles, nil
}

// parseFile parses a workspace configuration file.
func (w *workspace) parseFile(filename string) (*workspaceConfig, error) {
	w.verbosePrint("Parsing workspace configuration...")

	// Read workspace file
	content, err := w.fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWorkspaceFileReadError, err)
	}

	// Parse JSON
	var config workspaceConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, ErrWorkspaceFileMalformed
	}

	// Validate folders array
	if config.Folders == nil {
		return nil, ErrWorkspaceEmptyFolders
	}

	// Filter out null values and validate structure
	var validFolders []workspaceFolder
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

// getName extracts the workspace name from configuration or filename.
func (w *workspace) getName(config *workspaceConfig, filename string) string {
	// First try to get name from workspace configuration
	if config.Name != "" {
		return config.Name
	}

	// Fallback to filename without extension
	return strings.TrimSuffix(filepath.Base(filename), ".code-workspace")
}

// HandleMultipleFiles handles the selection of workspace files when multiple are found.
func (w *workspace) HandleMultipleFiles(workspaceFiles []string) (string, error) {
	w.verbosePrint(fmt.Sprintf("Multiple workspace files found: %d", len(workspaceFiles)))

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
		return w.HandleMultipleFiles(workspaceFiles)
	}

	w.verbosePrint(fmt.Sprintf("Selected workspace file: %s", selectedFile))
	return selectedFile, nil
}

// Validate validates all repositories in a workspace.
func (w *workspace) Validate() error {
	if w.verbose {
		w.logger.Logf("Validating workspace: %s", w.originalFile)
	}

	workspaceConfig, err := w.parseFile(w.originalFile)
	if err != nil {
		if w.verbose {
			w.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %w", ErrWorkspaceFileRead, err)
	}

	// Get workspace file directory for resolving relative paths
	workspaceDir := filepath.Dir(w.originalFile)

	for _, folder := range workspaceConfig.Folders {
		if err := w.validateRepository(folder, workspaceDir); err != nil {
			return err
		}
	}

	return nil
}

// validateRepository validates a single repository in a workspace.
func (w *workspace) validateRepository(folder workspaceFolder, workspaceDir string) error {
	// Resolve relative path from workspace file location
	resolvedPath := filepath.Join(workspaceDir, folder.Path)

	if w.verbose {
		w.logger.Logf("Validating repository: %s", resolvedPath)
	}

	if err := w.validateRepositoryPath(folder, resolvedPath); err != nil {
		return err
	}

	if err := w.validateRepositoryGit(folder, resolvedPath); err != nil {
		return err
	}

	// Validate Git configuration is functional
	err := w.validateGitConfiguration(resolvedPath)
	if err != nil {
		if w.verbose {
			w.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %s - %w", ErrInvalidRepositoryInWorkspace, folder.Path, err)
	}

	return nil
}

// validateRepositoryPath validates that the repository path exists.
func (w *workspace) validateRepositoryPath(folder workspaceFolder, resolvedPath string) error {
	// Check repository path exists
	exists, err := w.fs.Exists(resolvedPath)
	if err != nil {
		if w.verbose {
			w.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("repository not found in workspace: %s - %w", folder.Path, err)
	}

	if !exists {
		if w.verbose {
			w.logger.Logf("Error: repository path does not exist")
		}
		return fmt.Errorf("%w: %s", ErrRepositoryNotFoundInWorkspace, folder.Path)
	}

	return nil
}

// validateRepositoryGit validates that the repository has a .git directory and git status works.
func (w *workspace) validateRepositoryGit(folder workspaceFolder, resolvedPath string) error {
	// Verify path contains .git folder
	gitPath := filepath.Join(resolvedPath, ".git")
	exists, err := w.fs.Exists(gitPath)
	if err != nil {
		if w.verbose {
			w.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %s - %w", ErrInvalidRepositoryInWorkspace, folder.Path, err)
	}

	if !exists {
		if w.verbose {
			w.logger.Logf("Error: .git directory not found in repository")
		}
		return fmt.Errorf("%w: %s", ErrInvalidRepositoryInWorkspaceNoGit, folder.Path)
	}

	// Execute git status to ensure repository is working
	if w.verbose {
		w.logger.Logf("Executing git status in: %s", resolvedPath)
	}
	_, err = w.git.Status(resolvedPath)
	if err != nil {
		if w.verbose {
			w.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("%w: %s - %w", ErrInvalidRepositoryInWorkspace, folder.Path, err)
	}

	return nil
}

// validateGitConfiguration validates that Git is properly configured and working.
func (w *workspace) validateGitConfiguration(workDir string) error {
	if w.verbose {
		w.logger.Logf("Validating Git configuration in: %s", workDir)
	}

	// Execute git status to ensure basic Git functionality
	if w.verbose {
		w.logger.Logf("Executing git status in: %s", workDir)
	}
	_, err := w.git.Status(workDir)
	if err != nil {
		if w.verbose {
			w.logger.Logf("Error: %v", err)
		}
		return fmt.Errorf("git configuration error: %w", err)
	}

	return nil
}

// displaySelection displays the workspace selection prompt.
func (w *workspace) displaySelection(workspaceFiles []string) {
	fmt.Println("Multiple workspace files found. Please select one:")
	fmt.Println()

	for i, file := range workspaceFiles {
		fmt.Printf("%d. %s\n", i+1, file)
	}

	fmt.Println()
	fmt.Printf("Enter your choice (1-%d) or 'q' to quit: ", len(workspaceFiles))
}

// getUserSelection gets and validates user input for workspace selection.
func (w *workspace) getUserSelection(maxChoice int) (int, error) {
	return w.getUserSelectionWithRetries(maxChoice, 3)
}

// getUserSelectionWithRetries gets and validates user input with retry limit.
func (w *workspace) getUserSelectionWithRetries(maxChoice int, retries int) (int, error) {
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

// isQuitCommand checks if the input is a quit command.
func (w *workspace) isQuitCommand(input string) bool {
	quitCommands := []string{"q", "quit", "exit", "cancel"}
	for _, cmd := range quitCommands {
		if input == cmd {
			return true
		}
	}
	return false
}

// parseNumericInput parses numeric input from string.
func (w *workspace) parseNumericInput(input string) (int, error) {
	var choice int
	_, err := fmt.Sscanf(input, "%d", &choice)
	return choice, err
}

// isValidChoice checks if the choice is within valid range.
func (w *workspace) isValidChoice(choice, maxChoice int) bool {
	return choice >= 1 && choice <= maxChoice
}

// confirmSelection asks the user to confirm their workspace selection.
func (w *workspace) confirmSelection(workspaceFile string) (bool, error) {
	return w.confirmSelectionWithRetries(workspaceFile, 3)
}

// confirmSelectionWithRetries asks the user to confirm their workspace selection with retry limit.
func (w *workspace) confirmSelectionWithRetries(workspaceFile string, retries int) (bool, error) {
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

// parseConfirmationInput parses confirmation input.
func (w *workspace) parseConfirmationInput(input string) (bool, error) {
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

// Load handles the complete workspace loading workflow.
// It detects workspace files, handles user selection if multiple files are found,
// and loads the workspace configuration for display.
func (w *workspace) Load() error {
	// Detect workspace files
	workspaceFiles, err := w.detectWorkspaceFiles()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWorkspaceDetection, err)
	}

	if len(workspaceFiles) == 0 {
		w.originalFile = ""
		return nil
	}

	// If only one workspace file, store it directly
	if len(workspaceFiles) == 1 {
		w.originalFile = workspaceFiles[0]
	} else {
		// If multiple workspace files, handle user selection
		selectedFile, err := w.HandleMultipleFiles(workspaceFiles)
		if err != nil {
			return err
		}
		w.originalFile = selectedFile
	}

	// Load and display workspace configuration
	workspaceConfig, err := w.parseFile(w.originalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	w.verbosePrint("Workspace mode detected")

	workspaceName := w.getName(workspaceConfig, w.originalFile)
	w.verbosePrint(fmt.Sprintf("Found workspace: %s", workspaceName))

	if w.verbose {
		w.verbosePrint("Workspace configuration:")
		w.verbosePrint(fmt.Sprintf("  Folders: %d", len(workspaceConfig.Folders)))
		for _, folder := range workspaceConfig.Folders {
			w.verbosePrint(fmt.Sprintf("    - %s: %s", folder.Name, folder.Path))
		}
	}

	return nil
}

// verbosePrint prints a message only in verbose mode.
func (w *workspace) verbosePrint(message string) {
	if w.verbose {
		w.logger.Logf(message)
	}
}
