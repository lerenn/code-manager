package cm

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/issue"
	"github.com/lerenn/cm/pkg/logger"
	"github.com/lerenn/cm/pkg/status"
)

// WorkspaceConfig represents a VS Code-like workspace configuration.
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

// WorkspaceCreateWorktreeOpts contains optional parameters for CreateWorktree.
type WorkspaceCreateWorktreeOpts struct {
	IssueInfo *issue.Info
}

// workspace represents a workspace and provides methods for workspace operations.
type workspace struct {
	*base
	originalFile string
}

// newWorkspace creates a new Workspace instance.
func newWorkspace(
	fs fs.FS,
	git git.Git,
	config *config.Config,
	statusManager status.Manager,
	logger logger.Logger,
	verbose bool,
) *workspace {
	return &workspace{
		base: newBase(fs, git, config, statusManager, logger, verbose),
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

	w.verbosePrint("Found %d workspace file(s)", len(workspaceFiles))
	return workspaceFiles, nil
}

// parseFile parses a workspace configuration file.
func (w *workspace) parseFile(filename string) (*WorkspaceConfig, error) {
	w.verbosePrint("Parsing workspace configuration...")

	// Read workspace file
	content, err := w.fs.ReadFile(filename)
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

// getName extracts the workspace name from configuration or filename.
func (w *workspace) getName(config *WorkspaceConfig, filename string) string {
	// First try to get name from workspace configuration
	if config.Name != "" {
		return config.Name
	}

	// Fallback to filename without extension
	return strings.TrimSuffix(filepath.Base(filename), ".code-workspace")
}

// HandleMultipleFiles handles the selection of workspace files when multiple are found.
func (w *workspace) HandleMultipleFiles(workspaceFiles []string) (string, error) {
	w.verbosePrint("Multiple workspace files found: %d", len(workspaceFiles))

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

	w.verbosePrint("Selected workspace file: %s", selectedFile)
	return selectedFile, nil
}

// Validate validates all repositories in a workspace.
func (w *workspace) Validate() error {
	w.verbosePrint("Validating workspace: %s", w.originalFile)

	workspaceConfig, err := w.parseFile(w.originalFile)
	if err != nil {
		w.verbosePrint("Error: %v", err)
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
func (w *workspace) validateRepository(folder WorkspaceFolder, workspaceDir string) error {
	// Resolve relative path from workspace file location
	resolvedPath := filepath.Join(workspaceDir, folder.Path)

	w.verbosePrint("Validating repository: %s", resolvedPath)

	if err := w.validateRepositoryPath(folder, resolvedPath); err != nil {
		return err
	}

	if err := w.validateRepositoryGit(folder, resolvedPath); err != nil {
		return err
	}

	// Validate Git configuration is functional
	err := w.validateGitConfiguration(resolvedPath)
	if err != nil {
		w.verbosePrint("Error: %v", err)
		return fmt.Errorf("%w: %s - %w", ErrInvalidRepositoryInWorkspace, folder.Path, err)
	}

	return nil
}

// validateRepositoryPath validates that the repository path exists.
func (w *workspace) validateRepositoryPath(folder WorkspaceFolder, resolvedPath string) error {
	// Check repository path exists
	exists, err := w.fs.Exists(resolvedPath)
	if err != nil {
		w.verbosePrint("Error: %v", err)
		return fmt.Errorf("repository not found in workspace: %s - %w", folder.Path, err)
	}

	if !exists {
		w.verbosePrint("Error: repository path does not exist")
		return fmt.Errorf("%w: %s", ErrRepositoryNotFoundInWorkspace, folder.Path)
	}

	return nil
}

// validateRepositoryGit validates that the repository has a .git directory and git status works.
func (w *workspace) validateRepositoryGit(folder WorkspaceFolder, resolvedPath string) error {
	// Verify path contains .git folder
	gitPath := filepath.Join(resolvedPath, ".git")
	exists, err := w.fs.Exists(gitPath)
	if err != nil {
		w.verbosePrint("Error: %v", err)
		return fmt.Errorf("%w: %s - %w", ErrInvalidRepositoryInWorkspace, folder.Path, err)
	}

	if !exists {
		w.verbosePrint("Error: .git directory not found in repository")
		return fmt.Errorf("%w: %s", ErrInvalidRepositoryInWorkspaceNoGit, folder.Path)
	}

	// Execute git status to ensure repository is working
	w.verbosePrint("Executing git status in: %s", resolvedPath)
	_, err = w.git.Status(resolvedPath)
	if err != nil {
		w.verbosePrint("Error: %v", err)
		return fmt.Errorf("%w: %s - %w", ErrInvalidRepositoryInWorkspace, folder.Path, err)
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
	w.verbosePrint("Found workspace: %s", workspaceName)

	w.verbosePrint("Workspace configuration:")
	w.verbosePrint("  Folders: %d", len(workspaceConfig.Folders))
	for _, folder := range workspaceConfig.Folders {
		w.verbosePrint("    - %s: %s", folder.Name, folder.Path)
	}

	return nil
}

// ListWorktrees lists worktrees for workspace mode.
func (w *workspace) ListWorktrees() ([]status.Repository, error) {
	w.verbosePrint("Listing worktrees for workspace mode")

	// Load workspace configuration (only if not already loaded)
	if w.originalFile == "" {
		if err := w.Load(); err != nil {
			return nil, fmt.Errorf("failed to load workspace: %w", err)
		}
	}

	// Get all worktrees from status file
	allWorktrees, err := w.statusManager.ListAllWorktrees()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Convert workspace path to absolute for comparison with status file
	workspacePath, err := filepath.Abs(w.originalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for workspace file: %w", err)
	}

	w.verbosePrint("Looking for worktrees with workspace path: %s", workspacePath)
	w.verbosePrint("Total worktrees available: %d", len(allWorktrees))

	// Filter worktrees for this workspace and add remote information
	var workspaceWorktrees []status.Repository
	for _, worktree := range allWorktrees {
		w.verbosePrint("Checking worktree: URL=%s, Workspace=%s", worktree.URL, worktree.Workspace)
		if worktree.Workspace == workspacePath {
			// Get the remote for this branch
			remote, err := w.git.GetBranchRemote(".", worktree.Branch)
			if err != nil {
				// If we can't determine the remote, use "origin" as default
				remote = defaultRemote
			}

			// Create a copy with remote information
			worktreeWithRemote := worktree
			worktreeWithRemote.Remote = remote
			workspaceWorktrees = append(workspaceWorktrees, worktreeWithRemote)
			w.verbosePrint("✓ Found matching worktree: %s with remote: %s", worktree.URL, remote)
		}
	}

	return workspaceWorktrees, nil
}

// CreateWorktree creates worktrees for all repositories in the workspace.
func (w *workspace) CreateWorktree(branch string, opts ...WorkspaceCreateWorktreeOpts) error {
	w.verbosePrint("Creating worktrees for branch: %s", branch)

	// 1. Load and validate workspace configuration (only if not already loaded)
	if w.originalFile == "" {
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
	var workspaceOpts *WorkspaceCreateWorktreeOpts
	if len(opts) > 0 {
		workspaceOpts = &opts[0]
	}
	if err := w.createWorktreesForWorkspace(branch, workspaceOpts); err != nil {
		return fmt.Errorf("failed to create worktrees: %w", err)
	}

	w.verbosePrint("Workspace worktree creation completed successfully")
	return nil
}

// validateWorkspaceForWorktreeCreation validates workspace state before worktree creation.
func (w *workspace) validateWorkspaceForWorktreeCreation(branch string) error {
	w.verbosePrint("Validating workspace for worktree creation")

	workspaceConfig, err := w.parseFile(w.originalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Get workspace file directory for resolving relative paths
	workspaceDir := filepath.Dir(w.originalFile)

	for i, folder := range workspaceConfig.Folders {
		w.verbosePrint("Validating repository %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

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
			w.verbosePrint("Branch %s does not exist in %s, will create from current branch", branch, folder.Path)
		}

		// Validate directory creation permissions
		worktreePath := w.buildWorktreePath(repoURL, branch)

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
func (w *workspace) createWorktreesForWorkspace(branch string, opts *WorkspaceCreateWorktreeOpts) error {
	w.verbosePrint("Creating worktrees for all repositories in workspace")

	workspaceConfig, err := w.parseFile(w.originalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}

	// Get workspace name for worktree-specific workspace file
	workspaceName := w.getName(workspaceConfig, w.originalFile)
	workspaceDir := filepath.Dir(w.originalFile)

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

	// 1. Update status file with worktree entries
	if err := w.prepareWorktreeStatusEntries(workspaceConfig, workspaceDir, branch, &createdWorktrees, opts); err != nil {
		return err
	}

	// 2. Create worktree-specific workspace file
	if err := w.createWorktreeWorkspaceFile(createWorktreeWorkspaceFileParams{
		WorkspaceConfig:       workspaceConfig,
		WorkspaceName:         workspaceName,
		Branch:                branch,
		WorktreeWorkspacePath: worktreeWorkspacePath,
	}); err != nil {
		// Cleanup status entries on failure
		w.cleanupFailedWorktrees(createdWorktrees)
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
		return err
	}

	return nil
}

// prepareWorktreeStatusEntries prepares status file entries for all worktrees.
func (w *workspace) prepareWorktreeStatusEntries(
	workspaceConfig *WorkspaceConfig,
	workspaceDir string,
	branch string,
	createdWorktrees *[]struct {
		repoURL string
		branch  string
		path    string
	},
	opts *WorkspaceCreateWorktreeOpts,
) error {
	for i, folder := range workspaceConfig.Folders {
		w.verbosePrint("Preparing worktree %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

		resolvedPath := filepath.Join(workspaceDir, folder.Path)
		repoURL, err := w.git.GetRepositoryName(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
		}

		// Convert workspace path to absolute for status file storage
		workspacePath, err := filepath.Abs(w.originalFile)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for workspace file: %w", err)
		}

		// Add to status file
		var issueInfo *issue.Info
		if opts != nil && opts.IssueInfo != nil {
			issueInfo = opts.IssueInfo
		}
		if err := w.statusManager.AddWorktree(status.AddWorktreeParams{
			RepoURL:       repoURL,
			Branch:        branch,
			WorktreePath:  resolvedPath,
			WorkspacePath: workspacePath,
			IssueInfo:     issueInfo,
		}); err != nil {
			return fmt.Errorf("failed to add worktree to status file: %w", err)
		}

		*createdWorktrees = append(*createdWorktrees, struct {
			repoURL string
			branch  string
			path    string
		}{repoURL: repoURL, branch: branch, path: resolvedPath})
	}

	return nil
}

// createWorktreeDirectories creates worktree directories and executes Git worktree commands.
func (w *workspace) createWorktreeDirectories(
	workspaceConfig *WorkspaceConfig,
	workspaceDir string,
	branch string,
	createdWorktrees []struct {
		repoURL string
		branch  string
		path    string
	},
	worktreeWorkspacePath string,
	opts *WorkspaceCreateWorktreeOpts,
) error {
	for i, folder := range workspaceConfig.Folders {
		w.verbosePrint("Creating worktree %d/%d: %s", i+1, len(workspaceConfig.Folders), folder.Path)

		if err := w.createSingleWorktree(createSingleWorktreeParams{
			Folder:                folder,
			WorkspaceDir:          workspaceDir,
			Branch:                branch,
			CreatedWorktrees:      createdWorktrees,
			WorktreeWorkspacePath: worktreeWorkspacePath,
			Opts:                  opts,
		}); err != nil {
			return err
		}

		w.verbosePrint("✓ Worktree created successfully for %s", folder.Path)
	}

	return nil
}

// createSingleWorktreeParams contains parameters for creating a single worktree.
type createSingleWorktreeParams struct {
	Folder           WorkspaceFolder
	WorkspaceDir     string
	Branch           string
	CreatedWorktrees []struct {
		repoURL string
		branch  string
		path    string
	}
	WorktreeWorkspacePath string
	Opts                  *WorkspaceCreateWorktreeOpts
}

// createSingleWorktree creates a single worktree for a folder.
func (w *workspace) createSingleWorktree(params createSingleWorktreeParams) error {
	resolvedPath := filepath.Join(params.WorkspaceDir, params.Folder.Path)
	repoURL, err := w.git.GetRepositoryName(resolvedPath)
	if err != nil {
		w.cleanupOnFailure(params.CreatedWorktrees, params.WorktreeWorkspacePath)
		return fmt.Errorf("failed to get repository URL for %s: %w", params.Folder.Path, err)
	}

	worktreePath := w.buildWorktreePath(repoURL, params.Branch)

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
			w.verbosePrint("Warning: failed to clean up worktree directory: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree for %s: %w", params.Folder.Path, err)
	}

	return nil
}

// cleanupOnFailure performs cleanup operations when worktree creation fails.
func (w *workspace) cleanupOnFailure(
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
func (w *workspace) ensureBranchExists(params ensureBranchExistsParams) error {
	// Check if branch exists
	exists, err := w.git.BranchExists(params.ResolvedPath, params.Branch)
	if err != nil {
		// Cleanup on failure
		w.cleanupFailedWorktrees(params.CreatedWorktrees)
		w.cleanupWorktreeWorkspaceFile(params.WorktreeWorkspacePath)
		return fmt.Errorf("failed to check branch existence for %s: %w", params.FolderPath, err)
	}

	if !exists {
		w.verbosePrint("Branch %s does not exist in %s, creating from current branch", params.Branch, params.FolderPath)
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
	WorkspaceConfig       *WorkspaceConfig
	WorkspaceName         string
	Branch                string
	WorktreeWorkspacePath string
}

// createWorktreeWorkspaceFile creates the worktree-specific workspace file.
func (w *workspace) createWorktreeWorkspaceFile(params createWorktreeWorkspaceFileParams) error {
	w.verbosePrint("Creating worktree-specific workspace file")

	// Ensure workspaces directory exists
	workspacesDir := filepath.Dir(params.WorktreeWorkspacePath)
	if err := w.fs.MkdirAll(workspacesDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspaces directory: %w", err)
	}

	// Sanitize branch name for workspace name (replace slashes with hyphens)
	sanitizedBranchForName := strings.ReplaceAll(params.Branch, "/", "-")

	// Create worktree workspace configuration
	worktreeConfig := struct {
		Name    string            `json:"name,omitempty"`
		Folders []WorkspaceFolder `json:"folders"`
	}{
		Name:    fmt.Sprintf("%s-%s", params.WorkspaceName, sanitizedBranchForName),
		Folders: make([]WorkspaceFolder, len(params.WorkspaceConfig.Folders)),
	}

	// Update folder paths to point to worktree directories
	for i, folder := range params.WorkspaceConfig.Folders {
		// Get repository URL for this folder
		resolvedPath := filepath.Join(filepath.Dir(w.originalFile), folder.Path)
		repoURL, err := w.git.GetRepositoryName(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to get repository URL for %s: %w", folder.Path, err)
		}

		worktreeConfig.Folders[i] = WorkspaceFolder{
			Name: folder.Name,
			Path: w.buildWorktreePath(repoURL, params.Branch),
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

	w.verbosePrint("Worktree workspace file created: %s", params.WorktreeWorkspacePath)
	return nil
}

// cleanupFailedWorktrees removes worktree entries from status file.
func (w *workspace) cleanupFailedWorktrees(createdWorktrees []struct {
	repoURL string
	branch  string
	path    string
}) {
	w.verbosePrint("Cleaning up failed worktrees from status file")
	for _, worktree := range createdWorktrees {
		if err := w.statusManager.RemoveWorktree(worktree.repoURL, worktree.branch); err != nil {
			w.verbosePrint("Warning: failed to remove worktree from status file: %v", err)
		}
	}
}

// cleanupWorktreeWorkspaceFile removes the worktree-specific workspace file.
func (w *workspace) cleanupWorktreeWorkspaceFile(worktreeWorkspacePath string) {
	w.verbosePrint("Cleaning up worktree workspace file")
	if err := w.fs.RemoveAll(worktreeWorkspacePath); err != nil {
		w.verbosePrint("Warning: failed to remove worktree workspace file: %v", err)
	}
}

// DeleteWorktree deletes worktrees for the workspace with the specified branch.
func (w *workspace) DeleteWorktree(branch string, force bool) error {
	w.verbosePrint("Deleting worktrees for branch: %s", branch)

	// Load workspace configuration (only if not already loaded)
	if w.originalFile == "" {
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
	workspaceConfig, err := w.parseFile(w.originalFile)
	if err != nil {
		return fmt.Errorf("failed to parse workspace file: %w", err)
	}
	workspaceName := w.getName(workspaceConfig, w.originalFile)

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
		w.verbosePrint("Warning: failed to remove worktree workspace file: %v", err)
	}

	// Remove worktree entries from status file
	if err := w.removeWorktreeStatusEntries(workspaceWorktrees, force); err != nil {
		return err
	}

	w.verbosePrint("Workspace worktree deletion completed successfully")
	return nil
}

// getWorkspaceWorktrees gets all worktrees for this workspace and branch.
func (w *workspace) getWorkspaceWorktrees(branch string) ([]status.Repository, error) {
	// Get all worktrees for this workspace and branch
	allWorktrees, err := w.statusManager.ListAllWorktrees()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	// Convert workspace path to absolute for comparison with status file
	workspacePath, err := filepath.Abs(w.originalFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for workspace file: %w", err)
	}

	w.verbosePrint("Looking for worktrees with workspace path: %s", workspacePath)
	w.verbosePrint("Total worktrees available: %d", len(allWorktrees))

	// Filter worktrees for this workspace and branch
	var workspaceWorktrees []status.Repository
	for _, worktree := range allWorktrees {
		w.verbosePrint("Checking worktree: URL=%s, Workspace=%s, Branch=%s",
			worktree.URL, worktree.Workspace, worktree.Branch)
		if worktree.Workspace == workspacePath && worktree.Branch == branch {
			workspaceWorktrees = append(workspaceWorktrees, worktree)
			w.verbosePrint("✓ Found matching worktree: %s", worktree.URL)
		}
	}

	return workspaceWorktrees, nil
}

// deleteWorktreeRepositories deletes worktrees for all repositories.
func (w *workspace) deleteWorktreeRepositories(workspaceWorktrees []status.Repository, force bool) error {
	for i, worktree := range workspaceWorktrees {
		w.verbosePrint("Deleting worktree %d/%d: %s", i+1, len(workspaceWorktrees), worktree.URL)

		// Get original repository path
		originalRepoPath := worktree.Path

		// Delete Git worktree
		worktreePath := w.buildWorktreePath(worktree.URL, worktree.Branch)
		if err := w.git.RemoveWorktree(originalRepoPath, worktreePath); err != nil {
			if !force {
				return fmt.Errorf("failed to delete Git worktree for %s: %w", worktree.URL, err)
			}
			w.verbosePrint("Warning: failed to delete Git worktree for %s: %v", worktree.URL, err)
		}

		// Remove worktree directory
		if err := w.fs.RemoveAll(worktreePath); err != nil {
			if !force {
				return fmt.Errorf("failed to remove worktree directory %s: %w", worktreePath, err)
			}
			w.verbosePrint("Warning: failed to remove worktree directory %s: %v", worktreePath, err)
		}

		w.verbosePrint("✓ Worktree deleted successfully for %s", worktree.URL)
	}

	return nil
}

// removeWorktreeStatusEntries removes worktree entries from status file.
func (w *workspace) removeWorktreeStatusEntries(workspaceWorktrees []status.Repository, force bool) error {
	for _, worktree := range workspaceWorktrees {
		if err := w.statusManager.RemoveWorktree(worktree.URL, worktree.Branch); err != nil {
			if !force {
				return fmt.Errorf("failed to remove worktree from status file: %w", err)
			}
			w.verbosePrint("Warning: failed to remove worktree from status file: %v", err)
		}
	}

	return nil
}
