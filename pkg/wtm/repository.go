package wtm

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/forge"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/logger"
	"github.com/lerenn/wtm/pkg/status"
)

const defaultRemote = "origin"

// CreateWorktreeOpts contains optional parameters for CreateWorktree.
type CreateWorktreeOpts struct {
	IssueInfo *forge.IssueInfo
}

// repository represents a single Git repository and provides methods for repository operations.
type repository struct {
	*base
}

// newRepository creates a new Repository instance.
func newRepository(
	fs fs.FS,
	git git.Git,
	config *config.Config,
	statusManager status.Manager,
	logger logger.Logger,
	verbose bool,
) *repository {
	return &repository{
		base: newBase(fs, git, config, statusManager, logger, verbose),
	}
}

// Validate validates that the current directory is a working Git repository.
func (r *repository) Validate() error {
	r.verbosePrint("Validating repository: %s", ".")

	// Check if we're in a Git repository
	exists, err := r.IsGitRepository()
	if err != nil {
		return err
	}
	if !exists {
		return ErrGitRepositoryNotFound
	}

	if err := r.validateGitStatus(); err != nil {
		return err
	}

	// Validate Git configuration is functional
	return r.validateGitConfiguration(".")
}

// CreateWorktree creates a worktree for the repository with the specified branch.
func (r *repository) CreateWorktree(branch string, opts ...CreateWorktreeOpts) error {
	r.verbosePrint("Creating worktree for single repository with branch: %s", branch)

	// Validate and prepare repository
	repoURL, worktreePath, err := r.prepareWorktreeCreation(branch)
	if err != nil {
		return err
	}

	// Create the worktree
	if err := r.executeWorktreeCreation(repoURL, branch, worktreePath); err != nil {
		return err
	}

	// Create initial commit with issue information if provided
	if len(opts) > 0 && opts[0].IssueInfo != nil {
		if err := r.createInitialCommitWithIssue(worktreePath, opts[0].IssueInfo); err != nil {
			return fmt.Errorf("failed to create initial commit: %w", err)
		}
	}

	r.verbosePrint("Successfully created worktree for branch %s at %s", branch, worktreePath)

	return nil
}

// ListWorktrees lists all worktrees for the current repository.
func (r *repository) ListWorktrees() ([]status.Repository, error) {
	r.verbosePrint("Listing worktrees for single repository mode")

	// Note: Repository validation is already done in mode detection, so we skip it here
	// to avoid duplicate validation calls

	// 1. Extract repository name from remote origin URL (fallback to local path if no remote)
	repoName, err := r.git.GetRepositoryName(".")
	if err != nil {
		return nil, fmt.Errorf("failed to get repository name: %w", err)
	}

	r.verbosePrint("Repository name: %s", repoName)

	// 2. Load all worktrees from status file
	allWorktrees, err := r.statusManager.ListAllWorktrees()
	if err != nil {
		return nil, fmt.Errorf("failed to load worktrees from status file: %w", err)
	}

	r.verbosePrint("Found %d total worktrees in status file", len(allWorktrees))

	// 3. Filter worktrees to only include those for the current repository and add remote information
	var filteredWorktrees []status.Repository
	for _, worktree := range allWorktrees {
		if worktree.URL == repoName {
			// Get the remote for this branch
			remote, err := r.git.GetBranchRemote(".", worktree.Branch)
			if err != nil {
				// If we can't determine the remote, use "origin" as default
				remote = defaultRemote
			}

			// Create a copy with remote information
			worktreeWithRemote := worktree
			worktreeWithRemote.Remote = remote
			filteredWorktrees = append(filteredWorktrees, worktreeWithRemote)
		}
	}

	r.verbosePrint("Found %d worktrees for current repository", len(filteredWorktrees))

	return filteredWorktrees, nil
}

// IsGitRepository checks if the current directory is a Git repository (including worktrees).
func (r *repository) IsGitRepository() (bool, error) {
	r.verbosePrint("Checking if current directory is a Git repository...")

	// Check if .git exists
	exists, err := r.fs.Exists(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git existence: %w", err)
	}

	if !exists {
		r.verbosePrint("No .git found")
		return false, nil
	}

	// Check if .git is a directory (regular repository)
	isDir, err := r.fs.IsDir(".git")
	if err != nil {
		return false, fmt.Errorf("failed to check .git directory: %w", err)
	}

	if isDir {
		r.verbosePrint("Git repository detected (.git directory)")
		return true, nil
	}

	// If .git is not a directory, it must be a file (worktree)
	// Validate that it's actually a Git worktree file by checking for 'gitdir:' prefix
	r.verbosePrint("Checking if .git file is a valid worktree file...")

	content, err := r.fs.ReadFile(".git")
	if err != nil {
		r.verbosePrint("Failed to read .git file: %v", err)
		return false, nil
	}

	contentStr := strings.TrimSpace(string(content))
	if !strings.HasPrefix(contentStr, "gitdir:") {
		r.verbosePrint(".git file exists but is not a valid worktree file (missing 'gitdir:' prefix)")
		return false, nil
	}

	r.verbosePrint("Git worktree detected (.git file)")
	return true, nil
}

// IsWorkspaceFile checks if the current directory contains workspace files.
func (r *repository) IsWorkspaceFile() (bool, error) {
	r.verbosePrint("Checking for workspace files...")

	// Check for .code-workspace files
	workspaceFiles, err := r.fs.Glob("*.code-workspace")
	if err != nil {
		return false, fmt.Errorf("failed to check for workspace files: %w", err)
	}

	if len(workspaceFiles) > 0 {
		r.verbosePrint("Workspace files found: %v", workspaceFiles)
		return true, nil
	}

	r.verbosePrint("No workspace files found")
	return false, nil
}

// validateGitStatus validates that git status works in the repository.
func (r *repository) validateGitStatus() error {
	// Execute git status to ensure repository is working
	r.verbosePrint("Executing git status in: %s", ".")
	_, err := r.git.Status(".")
	if err != nil {
		r.verbosePrint("Error: %v", err)
		return fmt.Errorf("%w: %w", ErrGitRepositoryInvalid, err)
	}

	return nil
}

// prepareWorktreeCreation validates the repository and prepares the worktree path.
func (r *repository) prepareWorktreeCreation(branch string) (string, string, error) {
	// Validate repository
	repoURL, err := r.validateRepository(branch)
	if err != nil {
		return "", "", err
	}

	// Prepare worktree path
	worktreePath, err := r.prepareWorktreePath(repoURL, branch)
	if err != nil {
		return "", "", err
	}

	return repoURL, worktreePath, nil
}

// validateRepository validates the repository and gets the repository name.
func (r *repository) validateRepository(branch string) (string, error) {
	// Get current working directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Validate that we're in a Git repository
	isSingleRepo, err := r.IsGitRepository()
	if err != nil {
		return "", fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !isSingleRepo {
		return "", fmt.Errorf("current directory is not a Git repository")
	}

	// Get repository URL from remote origin URL with fallback to local path
	repoURL, err := r.git.GetRepositoryName(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	r.verbosePrint("Repository URL: %s", repoURL)

	// Check if worktree already exists in status file
	existingWorktree, err := r.statusManager.GetWorktree(repoURL, branch)
	if err == nil && existingWorktree != nil {
		return "", fmt.Errorf("%w for repository %s branch %s", ErrWorktreeExists, repoURL, branch)
	}

	// Validate repository state (placeholder for future validation)
	isClean, err := r.git.IsClean(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to check repository state: %w", err)
	}
	if !isClean {
		return "", fmt.Errorf("%w: repository is not in a clean state", ErrRepositoryNotClean)
	}

	return repoURL, nil
}

// prepareWorktreePath prepares the worktree directory path.
func (r *repository) prepareWorktreePath(repoURL, branch string) (string, error) {
	// Create worktree directory path
	worktreePath := r.buildWorktreePath(repoURL, branch)

	r.verbosePrint("Worktree path: %s", worktreePath)

	// Check if worktree directory already exists
	exists, err := r.fs.Exists(worktreePath)
	if err != nil {
		return "", fmt.Errorf("failed to check if worktree directory exists: %w", err)
	}
	if exists {
		return "", fmt.Errorf("%w: worktree directory already exists at %s", ErrDirectoryExists, worktreePath)
	}

	// Create worktree directory structure
	if err := r.fs.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory structure: %w", err)
	}

	return worktreePath, nil
}

// executeWorktreeCreation creates the branch and worktree.
func (r *repository) executeWorktreeCreation(repoURL, branch, worktreePath string) error {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Ensure branch exists
	if err := r.ensureBranchExists(currentDir, branch); err != nil {
		return err
	}

	// Create worktree with cleanup
	if err := r.createWorktreeWithCleanup(createWorktreeWithCleanupParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		CurrentDir:   currentDir,
	}); err != nil {
		return err
	}

	return nil
}

// ensureBranchExists ensures the branch exists, creating it if necessary.
func (r *repository) ensureBranchExists(currentDir, branch string) error {
	branchExists, err := r.git.BranchExists(currentDir, branch)
	if err != nil {
		return fmt.Errorf("failed to check if branch exists: %w", err)
	}

	if !branchExists {
		r.verbosePrint("Branch %s does not exist, creating from current branch", branch)
		if err := r.git.CreateBranch(currentDir, branch); err != nil {
			return fmt.Errorf("failed to create branch %s: %w", branch, err)
		}
	}

	return nil
}

// createWorktreeWithCleanupParams contains parameters for createWorktreeWithCleanup.
type createWorktreeWithCleanupParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	CurrentDir   string
}

// createWorktreeWithCleanup creates the worktree with proper cleanup on failure.
func (r *repository) createWorktreeWithCleanup(params createWorktreeWithCleanupParams) error {
	// Update status file with worktree entry (before creating the worktree for proper cleanup)
	// Store the original repository path, not the worktree path
	if err := r.statusManager.AddWorktree(status.AddWorktreeParams{
		RepoURL:       params.RepoURL,
		Branch:        params.Branch,
		WorktreePath:  params.CurrentDir,
		WorkspacePath: "",
	}); err != nil {
		// Clean up created directory on status update failure
		if cleanupErr := r.cleanupWorktreeDirectory(params.WorktreePath); cleanupErr != nil {
			r.logger.Logf("Warning: failed to clean up directory after status update failure: %v", cleanupErr)
		}
		return fmt.Errorf("failed to add worktree to status: %w", err)
	}

	// Create the Git worktree
	if err := r.git.CreateWorktree(params.CurrentDir, params.WorktreePath, params.Branch); err != nil {
		// Clean up on worktree creation failure
		if cleanupErr := r.statusManager.RemoveWorktree(params.RepoURL, params.Branch); cleanupErr != nil {
			r.logger.Logf("Warning: failed to remove worktree from status after creation failure: %v", cleanupErr)
		}
		if cleanupErr := r.cleanupWorktreeDirectory(params.WorktreePath); cleanupErr != nil {
			r.logger.Logf("Warning: failed to clean up directory after worktree creation failure: %v", cleanupErr)
		}
		return fmt.Errorf("failed to create Git worktree: %w", err)
	}

	return nil
}

// DeleteWorktree deletes a worktree for the repository with the specified branch.
func (r *repository) DeleteWorktree(branch string, force bool) error {
	r.verbosePrint("Deleting worktree for single repository with branch: %s", branch)

	// Validate and prepare worktree deletion
	repoURL, worktreePath, err := r.prepareWorktreeDeletion(branch)
	if err != nil {
		return err
	}

	// Execute the deletion
	if err := r.executeWorktreeDeletion(executeWorktreeDeletionParams{
		RepoURL:      repoURL,
		Branch:       branch,
		WorktreePath: worktreePath,
		Force:        force,
	}); err != nil {
		return err
	}

	r.verbosePrint("Successfully deleted worktree for branch %s", branch)

	return nil
}

// prepareWorktreeDeletion validates the repository and prepares the worktree deletion.
func (r *repository) prepareWorktreeDeletion(branch string) (string, string, error) {
	// Get current working directory
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Validate that we're in a Git repository
	isSingleRepo, err := r.IsGitRepository()
	if err != nil {
		return "", "", fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !isSingleRepo {
		return "", "", fmt.Errorf("current directory is not a Git repository")
	}

	// Get repository URL from remote origin URL with fallback to local path
	repoURL, err := r.git.GetRepositoryName(currentDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get repository URL: %w", err)
	}

	r.verbosePrint("Repository URL: %s", repoURL)

	// Check if worktree exists in status file
	existingWorktree, err := r.statusManager.GetWorktree(repoURL, branch)
	if err != nil || existingWorktree == nil {
		return "", "", fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotInStatus, repoURL, branch)
	}

	// Get worktree path from Git
	worktreePath, err := r.git.GetWorktreePath(currentDir, branch)
	if err != nil {
		return "", "", fmt.Errorf("failed to get worktree path: %w", err)
	}

	r.verbosePrint("Worktree path: %s", worktreePath)

	return repoURL, worktreePath, nil
}

// executeWorktreeDeletionParams contains parameters for executeWorktreeDeletion.
type executeWorktreeDeletionParams struct {
	RepoURL      string
	Branch       string
	WorktreePath string
	Force        bool
}

// executeWorktreeDeletion deletes the worktree with proper cleanup.
func (r *repository) executeWorktreeDeletion(params executeWorktreeDeletionParams) error {
	currentDir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Prompt for confirmation unless force flag is used
	if !params.Force {
		if err := r.promptForConfirmation(params.Branch, params.WorktreePath); err != nil {
			return err
		}
	}

	// Remove worktree from Git tracking first
	if err := r.git.RemoveWorktree(currentDir, params.WorktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree from Git: %w", err)
	}

	// Remove worktree directory
	if err := r.fs.RemoveAll(params.WorktreePath); err != nil {
		return fmt.Errorf("failed to remove worktree directory: %w", err)
	}

	// Remove entry from status file
	if err := r.statusManager.RemoveWorktree(params.RepoURL, params.Branch); err != nil {
		return fmt.Errorf("failed to remove worktree from status: %w", err)
	}

	return nil
}

// promptForConfirmation prompts the user for confirmation before deletion.
func (r *repository) promptForConfirmation(branch, worktreePath string) error {
	fmt.Printf("You are about to delete the worktree for branch '%s'\n", branch)
	fmt.Printf("Worktree path: %s\n", worktreePath)
	fmt.Print("Are you sure you want to continue? (y/N): ")

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	result, err := r.parseConfirmationInput(input)
	if err != nil {
		return err
	}

	if !result {
		return ErrDeletionCancelled
	}

	return nil
}

// LoadWorktree loads a branch from a remote source and creates a worktree.
func (r *repository) LoadWorktree(remoteSource, branchName string) error {
	r.verbosePrint("Loading branch: remote=%s, branch=%s", remoteSource, branchName)

	// 1. Validate current directory is a Git repository
	gitExists, err := r.IsGitRepository()
	if err != nil {
		return fmt.Errorf("failed to validate Git repository: %w", err)
	}
	if !gitExists {
		return ErrGitRepositoryNotFound
	}

	// 2. Validate origin remote exists and is a valid Git hosting service URL
	if err := r.validateOriginRemote(); err != nil {
		return err
	}

	// 3. Parse remote source (default to "origin" if not specified)
	if remoteSource == "" {
		remoteSource = defaultRemote
	}

	// 4. Handle remote management
	if err := r.handleRemoteManagement(remoteSource); err != nil {
		return err
	}

	// 5. Fetch from the remote
	r.verbosePrint("Fetching from remote '%s'", remoteSource)
	if err := r.git.FetchRemote(".", remoteSource); err != nil {
		return fmt.Errorf("%w: %w", git.ErrFetchFailed, err)
	}

	// 6. Validate branch exists on remote
	r.verbosePrint("Checking if branch '%s' exists on remote '%s'", branchName, remoteSource)
	exists, err := r.git.BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   ".",
		RemoteName: remoteSource,
		Branch:     branchName,
	})
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: branch '%s' not found on remote '%s'", git.ErrBranchNotFoundOnRemote, branchName, remoteSource)
	}

	// 7. Create worktree for the branch (using existing worktree creation logic directly)
	r.verbosePrint("Creating worktree for branch '%s'", branchName)
	return r.CreateWorktree(branchName)
}

// validateOriginRemote validates that the origin remote exists and is a valid Git hosting service URL.
func (r *repository) validateOriginRemote() error {
	r.verbosePrint("Validating origin remote")

	// Check if origin remote exists
	exists, err := r.git.RemoteExists(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to check origin remote: %w", err)
	}
	if !exists {
		return ErrOriginRemoteNotFound
	}

	// Get origin remote URL
	originURL, err := r.git.GetRemoteURL(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote URL: %w", err)
	}

	// Validate that it's a valid Git hosting service URL
	if r.extractHostFromURL(originURL) == "" {
		return ErrOriginRemoteInvalidURL
	}

	return nil
}

// extractHostFromURL extracts the host from a Git remote URL.
func (r *repository) extractHostFromURL(url string) string {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle SSH format: git@host:user/repo
	if strings.Contains(url, "@") && strings.Contains(url, ":") {
		parts := strings.Split(url, ":")
		if len(parts) == 2 {
			hostParts := strings.Split(parts[0], "@")
			if len(hostParts) == 2 {
				return hostParts[1] // host
			}
		}
	}

	// Handle HTTPS format: https://host/user/repo
	if strings.HasPrefix(url, "http") {
		parts := strings.Split(url, "/")
		if len(parts) >= 3 {
			return parts[2] // host
		}
	}

	return ""
}

// handleRemoteManagement handles remote addition if the remote doesn't exist.
func (r *repository) handleRemoteManagement(remoteSource string) error {
	// If remote source is "origin", no need to add it
	if remoteSource == "origin" {
		r.verbosePrint("Using existing origin remote")
		return nil
	}

	// Check if remote already exists and handle existing remote
	if err := r.handleExistingRemote(remoteSource); err != nil {
		return err
	}

	// Add new remote
	return r.addNewRemote(remoteSource)
}

// handleExistingRemote checks if remote exists and handles it appropriately.
func (r *repository) handleExistingRemote(remoteSource string) error {
	exists, err := r.git.RemoteExists(".", remoteSource)
	if err != nil {
		return fmt.Errorf("failed to check if remote '%s' exists: %w", remoteSource, err)
	}

	if exists {
		remoteURL, err := r.git.GetRemoteURL(".", remoteSource)
		if err != nil {
			return fmt.Errorf("failed to get remote URL: %w", err)
		}

		r.verbosePrint("Using existing remote '%s' with URL: %s", remoteSource, remoteURL)
		return nil
	}

	return nil
}

// addNewRemote adds a new remote for the given remote source.
func (r *repository) addNewRemote(remoteSource string) error {
	r.verbosePrint("Adding new remote '%s'", remoteSource)

	// Get repository information
	repoName, err := r.git.GetRepositoryName(".")
	if err != nil {
		return fmt.Errorf("failed to get repository name: %w", err)
	}

	originURL, err := r.git.GetRemoteURL(".", "origin")
	if err != nil {
		return fmt.Errorf("failed to get origin remote URL: %w", err)
	}

	// Construct remote URL
	remoteURL, err := r.constructRemoteURL(originURL, remoteSource, repoName)
	if err != nil {
		return err
	}

	r.verbosePrint("Constructed remote URL: %s", remoteURL)

	// Add the remote
	if err := r.git.AddRemote(".", remoteSource, remoteURL); err != nil {
		return fmt.Errorf("%w: %w", git.ErrRemoteAddFailed, err)
	}

	return nil
}

// constructRemoteURL constructs the remote URL based on origin URL and remote source.
func (r *repository) constructRemoteURL(originURL, remoteSource, repoName string) (string, error) {
	protocol := r.determineProtocol(originURL)
	host := r.extractHostFromURL(originURL)

	if host == "" {
		return "", fmt.Errorf("failed to extract host from origin URL: %s", originURL)
	}

	repoNameShort := r.extractRepoNameFromFullPath(repoName)

	if protocol == "ssh" {
		return fmt.Sprintf("git@%s:%s/%s.git", host, remoteSource, repoNameShort), nil
	}
	return fmt.Sprintf("https://%s/%s/%s.git", host, remoteSource, repoNameShort), nil
}

// determineProtocol determines the protocol (https or ssh) from the origin URL.
func (r *repository) determineProtocol(originURL string) string {
	if strings.HasPrefix(originURL, "git@") || strings.HasPrefix(originURL, "ssh://") {
		return "ssh"
	}
	return "https"
}

// extractRepoNameFromFullPath extracts just the repository name from the full path.
func (r *repository) extractRepoNameFromFullPath(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-1] // Return the last part (repo name)
	}
	return fullPath
}

// createInitialCommitWithIssue creates an initial commit with issue information.
func (r *repository) createInitialCommitWithIssue(worktreePath string, issueInfo *forge.IssueInfo) error {
	r.verbosePrint("Creating initial commit with issue information for worktree: %s", worktreePath)

	// Create a README file with issue information
	readmeContent := fmt.Sprintf(`# Issue #%d: %s

%s

## Issue Details
- **URL**: %s
- **State**: %s
- **Repository**: %s/%s

This worktree was created from issue #%d.
`, issueInfo.Number, issueInfo.Title, issueInfo.Description, issueInfo.URL, issueInfo.State, issueInfo.Owner, issueInfo.Repository, issueInfo.Number)

	// Write README file
	readmePath := filepath.Join(worktreePath, "README.md")
	if err := r.fs.WriteFileAtomic(readmePath, []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to write README file: %w", err)
	}

	// Add the file to git
	if err := r.git.Add(worktreePath, "README.md"); err != nil {
		return fmt.Errorf("failed to add README to git: %w", err)
	}

	// Create the commit
	commitMessage := fmt.Sprintf("%s\n\nIssue: %s", issueInfo.Title, issueInfo.URL)
	if err := r.git.Commit(worktreePath, commitMessage); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	r.verbosePrint("Successfully created initial commit with issue information")
	return nil
}
