// Package workspace provides workspace management functionality for CM.
package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
)

// Config represents the configuration of a workspace.
type Config struct {
	Name    string   `json:"name,omitempty"`
	Folders []Folder `json:"folders"`
}

// Folder represents a folder in a workspace.
type Folder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Workspace represents a workspace and provides methods for workspace operations.
type Workspace struct {
	fs            fs.FS
	git           git.Git
	config        *config.Config
	statusManager status.Manager
	logger        logger.Logger
	prompt        prompt.Prompt
	verbose       bool
	OriginalFile  string
}

// NewWorkspaceParams contains parameters for creating a new Workspace instance.
type NewWorkspaceParams struct {
	FS            fs.FS
	Git           git.Git
	Config        *config.Config
	StatusManager status.Manager
	Logger        logger.Logger
	Prompt        prompt.Prompt
	Verbose       bool
}

// NewWorkspace creates a new Workspace instance.
func NewWorkspace(params NewWorkspaceParams) *Workspace {
	return &Workspace{
		fs:            params.FS,
		git:           params.Git,
		config:        params.Config,
		statusManager: params.StatusManager,
		logger:        params.Logger,
		prompt:        params.Prompt,
		verbose:       params.Verbose,
	}
}

// verboseLogf prints a message if verbose mode is enabled.
func (w *Workspace) verboseLogf(format string, args ...interface{}) {
	if w.verbose {
		fmt.Printf(format+"\n", args...)
	}
}

// DetectWorkspaceFiles checks if the current directory contains workspace files.
func (w *Workspace) DetectWorkspaceFiles() ([]string, error) {
	w.verboseLogf("Checking for .code-workspace files...")

	// Check for workspace files
	workspaceFiles, err := w.fs.Glob("*.code-workspace")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToCheckWorkspaceFiles, err)
	}

	if len(workspaceFiles) == 0 {
		w.verboseLogf("No .code-workspace files found")
		return nil, nil
	}

	w.verboseLogf("Found %d workspace file(s)", len(workspaceFiles))
	return workspaceFiles, nil
}

// ParseFile parses a workspace configuration file.
func (w *Workspace) ParseFile(filename string) (*Config, error) {
	w.verboseLogf("Parsing workspace configuration...")

	// Read workspace file
	content, err := w.fs.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWorkspaceFileNotFound, err)
	}

	// Parse JSON
	var config Config
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, ErrInvalidWorkspaceFile
	}

	// Validate folders array
	if config.Folders == nil {
		return nil, ErrNoRepositoriesFound
	}

	// Filter out null values and validate structure
	var validFolders []Folder
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
		return nil, ErrNoRepositoriesFound
	}

	config.Folders = validFolders
	return &config, nil
}

// GetName extracts the workspace name from configuration or filename.
func (w *Workspace) GetName(config *Config, filename string) string {
	// First try to get name from workspace configuration
	if config.Name != "" {
		return config.Name
	}

	// Fallback to filename without extension
	return strings.TrimSuffix(filepath.Base(filename), ".code-workspace")
}

// HandleMultipleFiles handles the selection of workspace files when multiple are found.
func (w *Workspace) HandleMultipleFiles(workspaceFiles []string) (string, error) {
	w.verboseLogf("Multiple workspace files found: %d", len(workspaceFiles))

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

	w.verboseLogf("Selected workspace file: %s", selectedFile)
	return selectedFile, nil
}

// Validate validates all repositories in a workspace.
func (w *Workspace) Validate() error {
	w.verboseLogf("Validating workspace: %s", w.OriginalFile)

	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		w.verboseLogf("Error: %v", err)
		return fmt.Errorf("%w: %w", ErrWorkspaceFileNotFound, err)
	}

	// Get workspace file directory for resolving relative paths
	workspaceDir := filepath.Dir(w.OriginalFile)

	for _, folder := range workspaceConfig.Folders {
		if err := w.validateRepository(folder, workspaceDir); err != nil {
			return err
		}
	}

	return nil
}

// validateRepository validates a single repository in a workspace.
func (w *Workspace) validateRepository(folder Folder, workspaceDir string) error {
	// Resolve relative path from workspace file location
	resolvedPath := filepath.Join(workspaceDir, folder.Path)

	w.verboseLogf("Validating repository: %s", resolvedPath)

	if err := w.validateRepositoryPath(folder, resolvedPath); err != nil {
		return err
	}

	if err := w.validateRepositoryGit(folder, resolvedPath); err != nil {
		return err
	}

	// Validate Git configuration is functional
	err := w.validateGitConfiguration()
	if err != nil {
		w.verboseLogf("Error: %v", err)
		return fmt.Errorf("%w: %s - %w", ErrRepositoryNotFound, folder.Path, err)
	}

	return nil
}

// validateRepositoryPath validates that the repository path exists.
func (w *Workspace) validateRepositoryPath(folder Folder, resolvedPath string) error {
	// Check repository path exists
	exists, err := w.fs.Exists(resolvedPath)
	if err != nil {
		w.verboseLogf("Error: %v", err)
		return fmt.Errorf("repository not found in workspace: %s - %w", folder.Path, err)
	}

	if !exists {
		w.verboseLogf("Error: repository path does not exist")
		return fmt.Errorf("%w: %s", ErrRepositoryNotFound, folder.Path)
	}

	return nil
}

// validateRepositoryGit validates that the repository has a .git directory and git status works.
func (w *Workspace) validateRepositoryGit(folder Folder, resolvedPath string) error {
	// Verify path contains .git folder
	gitPath := filepath.Join(resolvedPath, ".git")
	exists, err := w.fs.Exists(gitPath)
	if err != nil {
		w.verboseLogf("Error: %v", err)
		return fmt.Errorf("%w: %s - %w", ErrRepositoryNotFound, folder.Path, err)
	}

	if !exists {
		w.verboseLogf("Error: .git directory not found in repository")
		return fmt.Errorf("%w: %s", ErrRepositoryNotFound, folder.Path)
	}

	// Execute git status to ensure repository is working
	w.verboseLogf("Executing git status in: %s", resolvedPath)
	_, err = w.git.Status(resolvedPath)
	if err != nil {
		w.verboseLogf("Error: %v", err)
		return fmt.Errorf("%w: %s - %w", ErrRepositoryNotFound, folder.Path, err)
	}

	return nil
}

// displaySelection displays the workspace selection prompt.
func (w *Workspace) displaySelection(workspaceFiles []string) {
	fmt.Println("Multiple workspace files found. Please select one:")
	fmt.Println()

	for i, file := range workspaceFiles {
		fmt.Printf("%d. %s\n", i+1, file)
	}

	fmt.Println()
	fmt.Printf("Enter your choice (1-%d) or 'q' to quit: ", len(workspaceFiles))
}

// getUserSelection gets and validates user input for workspace selection.
func (w *Workspace) getUserSelection(maxChoice int) (int, error) {
	return w.getUserSelectionWithRetries(maxChoice, 3)
}

// getUserSelectionWithRetries gets and validates user input with retry limit.
func (w *Workspace) getUserSelectionWithRetries(maxChoice int, retries int) (int, error) {
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
func (w *Workspace) confirmSelection(workspaceFile string) (bool, error) {
	return w.confirmSelectionWithRetries(workspaceFile, 3)
}

// confirmSelectionWithRetries asks the user to confirm their workspace selection with retry limit.
func (w *Workspace) confirmSelectionWithRetries(workspaceFile string, retries int) (bool, error) {
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

// Load handles the complete workspace loading workflow.
// It detects workspace files, handles user selection if multiple files are found,
// and loads the workspace configuration for display.
func (w *Workspace) Load() error {
	// If already loaded, just parse and display the configuration
	if w.OriginalFile != "" {
		workspaceConfig, err := w.ParseFile(w.OriginalFile)
		if err != nil {
			return fmt.Errorf("failed to parse workspace file: %w", err)
		}

		w.verboseLogf("Workspace mode detected")

		workspaceName := w.GetName(workspaceConfig, w.OriginalFile)
		w.verboseLogf("Found workspace: %s", workspaceName)

		w.verboseLogf("Workspace configuration:")
		w.verboseLogf("  Folders: %d", len(workspaceConfig.Folders))
		for _, folder := range workspaceConfig.Folders {
			w.verboseLogf("    - %s: %s", folder.Name, folder.Path)
		}

		return nil
	}

	// Detect workspace files
	workspaceFiles, err := w.DetectWorkspaceFiles()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWorkspaceFileNotFound, err)
	}

	if len(workspaceFiles) == 0 {
		w.OriginalFile = ""
		return nil
	}

	// If only one workspace file, store it directly
	if len(workspaceFiles) == 1 {
		w.OriginalFile = workspaceFiles[0]
	} else {
		// If multiple workspace files, handle user selection
		selectedFile, err := w.HandleMultipleFiles(workspaceFiles)
		if err != nil {
			return err
		}
		w.OriginalFile = selectedFile
	}

	// Load and display workspace configuration
	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	w.verboseLogf("Workspace mode detected")

	workspaceName := w.GetName(workspaceConfig, w.OriginalFile)
	w.verboseLogf("Found workspace: %s", workspaceName)

	w.verboseLogf("Workspace configuration:")
	w.verboseLogf("  Folders: %d", len(workspaceConfig.Folders))
	for _, folder := range workspaceConfig.Folders {
		w.verboseLogf("    - %s: %s", folder.Name, folder.Path)
	}

	return nil
}

// ListWorktrees lists worktrees for workspace mode.
func (w *Workspace) ListWorktrees() ([]status.WorktreeInfo, error) {
	w.verboseLogf("Listing worktrees for workspace mode")

	// Load workspace configuration (only if not already loaded)
	if w.OriginalFile == "" {
		if err := w.Load(); err != nil {
			return nil, fmt.Errorf("failed to load workspace: %w", err)
		}
	}

	// Get workspace path
	workspacePath, err := filepath.Abs(w.OriginalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for workspace file: %w", err)
	}

	// Get workspace from status
	workspace, err := w.statusManager.GetWorkspace(workspacePath)
	if err != nil {
		// If workspace not found, return empty list with no error
		if errors.Is(err, status.ErrWorkspaceNotFound) {
			return []status.WorktreeInfo{}, nil
		}
		return nil, err
	}

	// Get worktrees for each repository in the workspace
	var workspaceWorktrees []status.WorktreeInfo
	seenWorktrees := make(map[string]bool) // Track seen worktrees to avoid duplicates

	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := w.statusManager.GetRepository(repoURL)
		if err != nil {
			continue // Skip if repository not found
		}

		// Get worktrees for this repository
		for _, worktree := range repo.Worktrees {
			// Create a unique key for this worktree to avoid duplicates
			// Include repository URL to distinguish between worktrees from different repositories
			worktreeKey := fmt.Sprintf("%s:%s:%s", repoURL, worktree.Remote, worktree.Branch)
			if !seenWorktrees[worktreeKey] {
				workspaceWorktrees = append(workspaceWorktrees, worktree)
				seenWorktrees[worktreeKey] = true
			}
		}
	}

	return workspaceWorktrees, nil
}

// CreateWorktreeOpts contains optional parameters for worktree creation.
type CreateWorktreeOpts struct {
	IDEName string
}

// CreateWorktree creates worktrees for all repositories in the workspace.
func (w *Workspace) CreateWorktree(branch string, opts ...CreateWorktreeOpts) error {
	w.verboseLogf("Creating worktrees for branch: %s", branch)

	// 1. Load and validate workspace configuration (only if not already loaded)
	if w.OriginalFile == "" {
		if err := w.Load(); err != nil {
			return fmt.Errorf("failed to load workspace: %w", err)
		}
	}

	// 2. Validate all repositories in workspace
	if err := w.Validate(); err != nil {
		return fmt.Errorf("failed to validate workspace: %w", err)
	}

	// 3. Pre-validate worktree creation for all repositories
	if err := w.validateWorkspaceForWorktreeCreation(branch); err != nil {
		return fmt.Errorf("failed to validate workspace for worktree creation: %w", err)
	}

	// 4. Create worktrees for all repositories
	var workspaceOpts *CreateWorktreeOpts
	if len(opts) > 0 {
		workspaceOpts = &opts[0]
	}
	if err := w.createWorktreesForWorkspace(branch, workspaceOpts); err != nil {
		return fmt.Errorf("failed to create worktrees: %w", err)
	}

	w.verboseLogf("Workspace worktree creation completed successfully")
	return nil
}

// validateWorkspaceForWorktreeCreation validates workspace state before worktree creation.
func (w *Workspace) validateWorkspaceForWorktreeCreation(branch string) error {
	w.verboseLogf("Validating workspace for worktree creation")

	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Get workspace file directory for resolving relative paths
	workspaceDir := filepath.Dir(w.OriginalFile)

	for i, folder := range workspaceConfig.Folders {
		w.verboseLogf("Validating repository %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

		// Resolve relative path from workspace file location
		resolvedPath := filepath.Join(workspaceDir, folder.Path)

		// Check for existing worktrees in status file
		repoURL, err := w.git.GetRepositoryName(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
		}

		// Check if worktree already exists
		_, err = w.statusManager.GetWorktree(repoURL, branch)
		if err == nil {
			return fmt.Errorf("worktree already exists for repository %s branch %s", repoURL, branch)
		}

		// Check branch existence and create if needed
		exists, err := w.git.BranchExists(resolvedPath, branch)
		if err != nil {
			return fmt.Errorf("failed to check branch existence for %s: %w", folder.Path, err)
		}

		if !exists {
			w.verboseLogf("Branch %s does not exist in %s, will create from current branch", branch, folder.Path)
		}

		// Validate directory creation permissions
		worktreePath := w.buildWorktreePath(repoURL, "origin", branch)

		// Check if worktree directory already exists
		exists, err = w.fs.Exists(worktreePath)
		if err != nil {
			return fmt.Errorf("failed to check worktree directory existence: %w", err)
		}
		if exists {
			return fmt.Errorf("worktree directory already exists: %s", worktreePath)
		}
	}

	return nil
}

// createWorktreesForWorkspace creates worktrees for all repositories in the workspace.
func (w *Workspace) createWorktreesForWorkspace(branch string, opts *CreateWorktreeOpts) error {
	w.verboseLogf("Creating worktrees for all repositories in workspace")

	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Get workspace name for worktree-specific workspace file
	workspaceName := w.GetName(workspaceConfig, w.OriginalFile)
	workspaceDir := filepath.Dir(w.OriginalFile)

	// Track created worktrees for cleanup on failure
	var createdWorktrees []struct {
		repoURL string
		branch  string
		path    string
	}

	// Sanitize branch name for filename (replace slashes with hyphens)
	sanitizedBranchForFilename := strings.ReplaceAll(branch, "/", "-")

	// Create worktree-specific workspace file path
	worktreeWorkspacePath := filepath.Join(
		w.config.BasePath,
		"workspaces",
		fmt.Sprintf("%s-%s.code-workspace", workspaceName, sanitizedBranchForFilename),
	)

	// 1. Add workspace to status file if not already present
	if err := w.ensureWorkspaceInStatus(workspaceConfig, workspaceDir, workspaceName); err != nil {
		return err
	}

	// 2. Create worktree-specific workspace file
	if err := w.createWorktreeWorkspaceFile(createWorktreeWorkspaceFileParams{
		WorkspaceConfig:       workspaceConfig,
		WorkspaceName:         workspaceName,
		Branch:                branch,
		WorktreeWorkspacePath: worktreeWorkspacePath,
	}); err != nil {
		return fmt.Errorf("failed to create worktree workspace file: %w", err)
	}

	// 3. Create worktree directories and execute Git worktree commands
	if err := w.createWorktreeDirectories(
		workspaceConfig,
		workspaceDir,
		branch,
		createdWorktrees,
		worktreeWorkspacePath,
		opts,
	); err != nil {
		// Cleanup workspace file on failure
		if cleanupErr := w.fs.RemoveAll(worktreeWorkspacePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree workspace file: %v", cleanupErr)
		}
		return err
	}

	// 4. Update status file with worktree entries after successful creation
	if err := w.prepareWorktreeStatusEntries(workspaceConfig, workspaceDir, branch, &createdWorktrees, opts); err != nil {
		// Cleanup created worktrees and workspace file on failure
		w.cleanupFailedWorktrees(createdWorktrees)
		if cleanupErr := w.fs.RemoveAll(worktreeWorkspacePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree workspace file: %v", cleanupErr)
		}
		return err
	}

	return nil
}

// ensureWorkspaceInStatus ensures the workspace is added to the status file.
func (w *Workspace) ensureWorkspaceInStatus(workspaceConfig *Config, workspaceDir, workspaceName string) error {
	// Get workspace path
	workspacePath, err := filepath.Abs(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for workspace file: %w", err)
	}

	// Check if workspace already exists in status
	_, err = w.statusManager.GetWorkspace(workspacePath)
	if err == nil {
		// Workspace already exists, no need to add it
		return nil
	}

	// Collect repository URLs for the workspace
	var repoURLs []string
	for _, folder := range workspaceConfig.Folders {
		resolvedPath := filepath.Join(workspaceDir, folder.Path)
		repoURL, err := w.git.GetRepositoryName(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
		}
		repoURLs = append(repoURLs, repoURL)
	}

	// Add workspace to status file
	if err := w.statusManager.AddWorkspace(workspacePath, status.AddWorkspaceParams{
		Worktree:     workspaceName,
		Repositories: repoURLs,
	}); err != nil {
		return fmt.Errorf("failed to add workspace to status file: %w", err)
	}

	return nil
}

// prepareWorktreeStatusEntries prepares status file entries for all worktrees.
func (w *Workspace) prepareWorktreeStatusEntries(
	workspaceConfig *Config,
	workspaceDir string,
	branch string,
	createdWorktrees *[]struct {
		repoURL string
		branch  string
		path    string
	},
	_ *CreateWorktreeOpts,
) error {
	for i, folder := range workspaceConfig.Folders {
		w.verboseLogf("Preparing worktree %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

		resolvedPath := filepath.Join(workspaceDir, folder.Path)
		repoURL, err := w.git.GetRepositoryName(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
		}

		// Convert workspace path to absolute for status file storage
		workspacePath, err := filepath.Abs(w.OriginalFile)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for workspace file: %w", err)
		}

		// Add to status file
		if err := w.addWorktreeToStatus(repoURL, branch, resolvedPath, workspacePath); err != nil {
			return err
		}

		*createdWorktrees = append(*createdWorktrees, struct {
			repoURL string
			branch  string
			path    string
		}{repoURL: repoURL, branch: branch, path: resolvedPath})
	}

	return nil
}

// addWorktreeToStatus adds the worktree to the status file with proper error handling.
func (w *Workspace) addWorktreeToStatus(repoURL, branch, resolvedPath, workspacePath string) error {
	if err := w.statusManager.AddWorktree(status.AddWorktreeParams{
		RepoURL:       repoURL,
		Branch:        branch,
		WorktreePath:  resolvedPath,
		WorkspacePath: workspacePath,
		Remote:        "origin", // Set the remote to "origin" for workspace worktrees
	}); err != nil {
		return w.handleStatusAddError(err, repoURL, resolvedPath, branch, workspacePath)
	}
	return nil
}

// handleStatusAddError handles errors when adding worktree to status.
func (w *Workspace) handleStatusAddError(err error, repoURL, resolvedPath, branch, workspacePath string) error {
	// Check if the error is due to repository not found, and auto-add it
	if errors.Is(err, status.ErrRepositoryNotFound) {
		if addErr := w.autoAddRepositoryToStatus(repoURL, resolvedPath); addErr != nil {
			return fmt.Errorf("failed to auto-add repository to status: %w", addErr)
		}

		// Try adding the worktree again
		if err := w.statusManager.AddWorktree(status.AddWorktreeParams{
			RepoURL:       repoURL,
			Branch:        branch,
			WorktreePath:  resolvedPath,
			WorkspacePath: workspacePath,
			Remote:        "origin", // Set the remote to "origin" for workspace worktrees
		}); err != nil {
			return fmt.Errorf("failed to add worktree to status file: %w", err)
		}
	} else {
		return fmt.Errorf("failed to add worktree to status file: %w", err)
	}
	return nil
}

// createWorktreeDirectories creates worktree directories and executes Git worktree commands.
func (w *Workspace) createWorktreeDirectories(
	workspaceConfig *Config,
	workspaceDir string,
	branch string,
	createdWorktrees []struct {
		repoURL string
		branch  string
		path    string
	},
	worktreeWorkspacePath string,
	_ *CreateWorktreeOpts,
) error {
	for i, folder := range workspaceConfig.Folders {
		w.verboseLogf("Creating worktree %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

		if err := w.createSingleWorktree(createSingleWorktreeParams{
			Folder:                folder,
			WorkspaceDir:          workspaceDir,
			Branch:                branch,
			CreatedWorktrees:      createdWorktrees,
			WorktreeWorkspacePath: worktreeWorkspacePath,
			Opts:                  nil,
		}); err != nil {
			return err
		}

		w.verboseLogf("✓ Worktree created successfully for %s", folder.Path)
	}

	return nil
}

// createSingleWorktreeParams contains parameters for creating a single worktree.
type createSingleWorktreeParams struct {
	Folder           Folder
	WorkspaceDir     string
	Branch           string
	CreatedWorktrees []struct {
		repoURL string
		branch  string
		path    string
	}
	WorktreeWorkspacePath string
	Opts                  *CreateWorktreeOpts
}

// createSingleWorktree creates a single worktree for a folder.
func (w *Workspace) createSingleWorktree(params createSingleWorktreeParams) error {
	resolvedPath := filepath.Join(params.WorkspaceDir, params.Folder.Path)
	repoURL, err := w.git.GetRepositoryName(resolvedPath)
	if err != nil {
		w.cleanupOnFailure(params.CreatedWorktrees, params.WorktreeWorkspacePath)
		return fmt.Errorf("failed to get repository URL for %s: %w", params.Folder.Path, err)
	}

	worktreePath := w.buildWorktreePath(repoURL, "origin", params.Branch)

	// Ensure branch exists
	if err := w.ensureBranchExists(ensureBranchExistsParams{
		ResolvedPath:          resolvedPath,
		Branch:                params.Branch,
		FolderPath:            params.Folder.Path,
		CreatedWorktrees:      params.CreatedWorktrees,
		WorktreeWorkspacePath: params.WorktreeWorkspacePath,
	}); err != nil {
		return err
	}

	// Create worktree directory
	if err := w.fs.MkdirAll(worktreePath, 0755); err != nil {
		w.cleanupOnFailure(params.CreatedWorktrees, params.WorktreeWorkspacePath)
		return fmt.Errorf("failed to create worktree directory %s: %w", worktreePath, err)
	}

	// Execute Git worktree creation command
	if err := w.git.CreateWorktree(resolvedPath, worktreePath, params.Branch); err != nil {
		w.cleanupOnFailure(params.CreatedWorktrees, params.WorktreeWorkspacePath)
		if cleanupErr := w.cleanupWorktreeDirectory(worktreePath); cleanupErr != nil {
			w.verboseLogf("Warning: failed to clean up worktree directory: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree for %s: %w", params.Folder.Path, err)
	}

	return nil
}

// cleanupOnFailure performs cleanup operations when worktree creation fails.
func (w *Workspace) cleanupOnFailure(
	createdWorktrees []struct {
		repoURL string
		branch  string
		path    string
	},
	worktreeWorkspacePath string,
) {
	w.cleanupFailedWorktrees(createdWorktrees)
	w.cleanupWorktreeWorkspaceFile(worktreeWorkspacePath)
}

// ensureBranchExistsParams contains parameters for ensuring a branch exists.
type ensureBranchExistsParams struct {
	ResolvedPath     string
	Branch           string
	FolderPath       string
	CreatedWorktrees []struct {
		repoURL string
		branch  string
		path    string
	}
	WorktreeWorkspacePath string
}

// ensureBranchExists ensures that the specified branch exists in the repository.
func (w *Workspace) ensureBranchExists(params ensureBranchExistsParams) error {
	// Check if branch exists
	exists, err := w.git.BranchExists(params.ResolvedPath, params.Branch)
	if err != nil {
		// Cleanup on failure
		w.cleanupFailedWorktrees(params.CreatedWorktrees)
		w.cleanupWorktreeWorkspaceFile(params.WorktreeWorkspacePath)
		return fmt.Errorf("failed to check branch existence for %s: %w", params.FolderPath, err)
	}

	if !exists {
		w.verboseLogf("Branch %s does not exist in %s, creating from current branch", params.Branch, params.FolderPath)
		if err := w.git.CreateBranch(params.ResolvedPath, params.Branch); err != nil {
			// Cleanup on failure
			w.cleanupFailedWorktrees(params.CreatedWorktrees)
			w.cleanupWorktreeWorkspaceFile(params.WorktreeWorkspacePath)
			return fmt.Errorf("failed to create branch %s for %s: %w", params.Branch, params.FolderPath, err)
		}
	}

	return nil
}

// createWorktreeWorkspaceFileParams contains parameters for creating a worktree workspace file.
type createWorktreeWorkspaceFileParams struct {
	WorkspaceConfig       *Config
	WorkspaceName         string
	Branch                string
	WorktreeWorkspacePath string
}

// createWorktreeWorkspaceFile creates the worktree-specific workspace file.
func (w *Workspace) createWorktreeWorkspaceFile(params createWorktreeWorkspaceFileParams) error {
	w.verboseLogf("Creating worktree-specific workspace file")

	// Ensure workspaces directory exists
	workspacesDir := filepath.Dir(params.WorktreeWorkspacePath)
	if err := w.fs.MkdirAll(workspacesDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspaces directory: %w", err)
	}

	// Sanitize branch name for workspace name (replace slashes with hyphens)
	sanitizedBranchForName := strings.ReplaceAll(params.Branch, "/", "-")

	// Create worktree workspace configuration
	worktreeConfig := struct {
		Name    string   `json:"name,omitempty"`
		Folders []Folder `json:"folders"`
	}{
		Name:    fmt.Sprintf("%s-%s", params.WorkspaceName, sanitizedBranchForName),
		Folders: make([]Folder, len(params.WorkspaceConfig.Folders)),
	}

	// Update folder paths to point to worktree directories
	for i, folder := range params.WorkspaceConfig.Folders {
		// Get repository URL for this folder
		resolvedPath := filepath.Join(filepath.Dir(w.OriginalFile), folder.Path)
		repoURL, err := w.git.GetRepositoryName(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
		}

		worktreeConfig.Folders[i] = Folder{
			Name: folder.Name,
			Path: w.buildWorktreePath(repoURL, "origin", params.Branch),
		}
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(worktreeConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal worktree workspace config: %w", err)
	}

	// Write worktree workspace file
	if err := w.fs.WriteFileAtomic(params.WorktreeWorkspacePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write worktree workspace file: %w", err)
	}

	w.verboseLogf("Worktree workspace file created: %s", params.WorktreeWorkspacePath)
	return nil
}

// cleanupFailedWorktrees removes worktree entries from status file.
func (w *Workspace) cleanupFailedWorktrees(createdWorktrees []struct {
	repoURL string
	branch  string
	path    string
}) {
	w.verboseLogf("Cleaning up failed worktrees from status file")
	for _, worktree := range createdWorktrees {
		if err := w.statusManager.RemoveWorktree(worktree.repoURL, worktree.branch); err != nil {
			w.verboseLogf("Warning: failed to remove worktree from status file: %v", err)
		}
	}
}

// cleanupWorktreeWorkspaceFile removes the worktree-specific workspace file.
func (w *Workspace) cleanupWorktreeWorkspaceFile(worktreeWorkspacePath string) {
	w.verboseLogf("Cleaning up worktree workspace file")
	if err := w.fs.RemoveAll(worktreeWorkspacePath); err != nil {
		w.verboseLogf("Warning: failed to remove worktree workspace file: %v", err)
	}
}

// DeleteWorktree deletes worktrees for the workspace with the specified branch.
func (w *Workspace) DeleteWorktree(branch string, force bool) error {
	w.verboseLogf("Deleting worktrees for branch: %s", branch)

	// Load workspace configuration (only if not already loaded)
	if w.OriginalFile == "" {
		if err := w.Load(); err != nil {
			return fmt.Errorf("failed to load workspace: %w", err)
		}
	}

	// Get worktrees for this workspace and branch
	workspaceWorktrees, err := w.getWorkspaceWorktrees(branch)
	if err != nil {
		return err
	}

	if len(workspaceWorktrees) == 0 {
		return fmt.Errorf("no worktrees found for workspace branch %s", branch)
	}

	// Get workspace name for worktree-specific workspace file
	workspaceConfig, err := w.ParseFile(w.OriginalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}
	workspaceName := w.GetName(workspaceConfig, w.OriginalFile)

	// Sanitize branch name for filename (replace slashes with hyphens)
	sanitizedBranchForFilename := strings.ReplaceAll(branch, "/", "-")

	worktreeWorkspacePath := filepath.Join(
		w.config.BasePath,
		"workspaces",
		fmt.Sprintf("%s-%s.code-workspace", workspaceName, sanitizedBranchForFilename),
	)

	// Delete worktrees for all repositories
	if err := w.deleteWorktreeRepositories(workspaceWorktrees, force); err != nil {
		return err
	}

	// Delete worktree-specific workspace file
	if err := w.fs.RemoveAll(worktreeWorkspacePath); err != nil {
		if !force {
			return fmt.Errorf("failed to remove worktree workspace file: %w", err)
		}
		w.verboseLogf("Warning: failed to remove worktree workspace file: %v", err)
	}

	// Remove worktree entries from status file
	if err := w.removeWorktreeStatusEntries(workspaceWorktrees, force); err != nil {
		return err
	}

	w.verboseLogf("Workspace worktree deletion completed successfully")
	return nil
}

// WorktreeWithRepo represents a worktree with its associated repository information.
type WorktreeWithRepo struct {
	status.WorktreeInfo
	RepoURL  string
	RepoPath string
}

// getWorkspaceWorktrees gets all worktrees for this workspace and branch.
func (w *Workspace) getWorkspaceWorktrees(branch string) ([]WorktreeWithRepo, error) {
	// Get workspace path
	workspacePath, err := filepath.Abs(w.OriginalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for workspace file: %w", err)
	}

	// Get workspace from status
	workspace, err := w.statusManager.GetWorkspace(workspacePath)
	if err != nil {
		// If workspace not found, return empty list with no error
		if errors.Is(err, status.ErrWorkspaceNotFound) {
			return []WorktreeWithRepo{}, nil
		}
		return nil, err
	}

	w.verboseLogf("Looking for worktrees with workspace path: %s", workspacePath)
	w.verboseLogf("Workspace repositories: %v", workspace.Repositories)

	// Get worktrees for each repository in the workspace that match the branch
	var workspaceWorktrees []WorktreeWithRepo

	for _, repoURL := range workspace.Repositories {
		// Get repository to check its worktrees
		repo, err := w.statusManager.GetRepository(repoURL)
		if err != nil {
			continue // Skip if repository not found
		}

		// Get worktrees for this repository that match the branch
		for _, worktree := range repo.Worktrees {
			if worktree.Branch == branch {
				workspaceWorktrees = append(workspaceWorktrees, WorktreeWithRepo{
					WorktreeInfo: worktree,
					RepoURL:      repoURL,
					RepoPath:     repo.Path,
				})
				w.verboseLogf("✓ Found matching worktree: %s:%s for repository %s", worktree.Remote, worktree.Branch, repoURL)
			}
		}
	}

	return workspaceWorktrees, nil
}

// deleteWorktreeRepositories deletes worktrees for all repositories.
func (w *Workspace) deleteWorktreeRepositories(workspaceWorktrees []WorktreeWithRepo, force bool) error {
	for i, worktreeWithRepo := range workspaceWorktrees {
		w.verboseLogf("Deleting worktree %d/%d: %s:%s for repository %s", i+1, len(workspaceWorktrees),
			worktreeWithRepo.Remote, worktreeWithRepo.Branch, worktreeWithRepo.RepoURL)

		// Delete Git worktree
		worktreePath := w.buildWorktreePath(worktreeWithRepo.RepoURL, worktreeWithRepo.Remote, worktreeWithRepo.Branch)
		if err := w.git.RemoveWorktree(worktreeWithRepo.RepoPath, worktreePath); err != nil {
			if !force {
				return fmt.Errorf("failed to delete Git worktree for %s:%s: %w",
					worktreeWithRepo.Remote, worktreeWithRepo.Branch, err)
			}
			w.verboseLogf("Warning: failed to delete Git worktree for %s:%s: %v",
				worktreeWithRepo.Remote, worktreeWithRepo.Branch, err)
		}

		// Remove worktree directory
		if err := w.fs.RemoveAll(worktreePath); err != nil {
			if !force {
				return fmt.Errorf("failed to remove worktree directory %s: %w", worktreePath, err)
			}
			w.verboseLogf("Warning: failed to remove worktree directory %s: %v", worktreePath, err)
		}

		w.verboseLogf("✓ Worktree deleted successfully for %s:%s", worktreeWithRepo.Remote, worktreeWithRepo.Branch)
	}

	return nil
}

// removeWorktreeStatusEntries removes worktree entries from status file.
func (w *Workspace) removeWorktreeStatusEntries(workspaceWorktrees []WorktreeWithRepo, force bool) error {
	for _, worktreeWithRepo := range workspaceWorktrees {
		if err := w.statusManager.RemoveWorktree(worktreeWithRepo.RepoURL, worktreeWithRepo.Branch); err != nil {
			if !force {
				return fmt.Errorf("failed to remove worktree from status file: %w", err)
			}
			w.verboseLogf("Warning: failed to remove worktree from status file: %v", err)
		}
	}

	return nil
}

// validateGitConfiguration validates Git configuration for the workspace.
func (w *Workspace) validateGitConfiguration() error {
	// Git validation is not implemented in the interface, so we'll skip it for now
	// This method is kept for future implementation
	return nil
}

// isQuitCommand checks if the input is a quit command.
func (w *Workspace) isQuitCommand(input string) bool {
	return strings.ToLower(strings.TrimSpace(input)) == "q" ||
		strings.ToLower(strings.TrimSpace(input)) == "quit" ||
		strings.ToLower(strings.TrimSpace(input)) == "exit"
}

// parseNumericInput parses numeric input from user.
func (w *Workspace) parseNumericInput(input string) (int, error) {
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
func (w *Workspace) isValidChoice(choice, maxChoice int) bool {
	return choice >= 1 && choice <= maxChoice
}

// parseConfirmationInput parses confirmation input from user.
func (w *Workspace) parseConfirmationInput(input string) (bool, error) {
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

// buildWorktreePath builds the worktree path for a repository, remote name, and branch.
func (w *Workspace) buildWorktreePath(repoURL, remoteName, branch string) string {
	// Use computed worktrees directory with new structure: $base_path/<repo_url>/<remote_name>/<branch>
	worktreesBase := filepath.Join(w.config.BasePath, "worktrees")
	return filepath.Join(worktreesBase, repoURL, remoteName, branch)
}

// autoAddRepositoryToStatus automatically adds a repository to the status file.
func (w *Workspace) autoAddRepositoryToStatus(repoURL, repoPath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(repoPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if it's a Git repository
	exists, err := w.fs.Exists(filepath.Join(absPath, ".git"))
	if err != nil {
		return fmt.Errorf("failed to check .git existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("not a Git repository: .git directory not found")
	}

	// Get remotes information
	remotes := make(map[string]status.Remote)

	// Check for origin remote
	originURL, err := w.git.GetRemoteURL(absPath, "origin")
	if err == nil && originURL != "" {
		remotes["origin"] = status.Remote{
			DefaultBranch: "main", // Default to main, could be enhanced to detect actual default branch
		}
	}

	// Add the repository to status
	if err := w.statusManager.AddRepository(repoURL, status.AddRepositoryParams{
		Path:    absPath,
		Remotes: remotes,
	}); err != nil {
		return fmt.Errorf("failed to add repository to status: %w", err)
	}

	return nil
}

// cleanupWorktreeDirectory cleans up the worktree directory.
func (w *Workspace) cleanupWorktreeDirectory(worktreePath string) error {
	if err := w.fs.RemoveAll(worktreePath); err != nil {
		return fmt.Errorf("failed to cleanup worktree directory: %w", err)
	}
	return nil
}
