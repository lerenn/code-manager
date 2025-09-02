// Package workspace provides workspace management functionality for CM.
package workspace

import (
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
	"github.com/lerenn/code-manager/pkg/worktree"
)

//go:generate mockgen -source=workspace.go -destination=mocks/workspace.gen.go -package=mocks

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

// Workspace interface provides workspace management capabilities.
type Workspace interface {
	// Load handles the complete workspace loading workflow.
	Load(force bool) error

	// CreateWorktree creates worktrees for all repositories in the workspace.
	CreateWorktree(branch string, force bool, opts ...CreateWorktreeOpts) (string, error)

	// DeleteWorktree deletes a worktree for the workspace with the specified branch.
	DeleteWorktree(branch string, force bool) error

	// ListWorktrees lists worktrees for workspace mode.
	ListWorktrees(force bool) ([]status.WorktreeInfo, error)

	// Validate validates all repositories in a workspace.
	Validate() error

	// DetectWorkspaceFiles checks if the current directory contains workspace files.
	DetectWorkspaceFiles() ([]string, error)

	// ParseFile parses a workspace configuration file.
	ParseFile(filename string) (Config, error)

	// GetName extracts the workspace name from configuration or filename.
	GetName(config Config, filename string) string

	// HandleMultipleFiles handles the selection of workspace files when multiple are found.
	HandleMultipleFiles(workspaceFiles []string, force bool) (string, error)

	// ValidateWorkspaceReferences validates that workspace references point to existing worktrees and repositories.
	ValidateWorkspaceReferences() error

	// SetLogger sets the logger for this workspace instance.
	SetLogger(logger logger.Logger)
}

// WorktreeProvider is a function type that creates worktree instances.
type WorktreeProvider func(params worktree.NewWorktreeParams) worktree.Worktree

// realWorkspace represents a workspace and provides methods for workspace operations.
type realWorkspace struct {
	fs               fs.FS
	git              git.Git
	config           config.Config
	statusManager    status.Manager
	logger           logger.Logger
	prompt           prompt.Prompter
	worktreeProvider WorktreeProvider
	OriginalFile     string
}

// NewWorkspaceParams contains parameters for creating a new Workspace instance.
type NewWorkspaceParams struct {
	FS               fs.FS
	Git              git.Git
	Config           config.Config
	StatusManager    status.Manager
	Logger           logger.Logger
	Prompt           prompt.Prompter
	WorktreeProvider WorktreeProvider
}

// NewWorkspace creates a new Workspace instance.
func NewWorkspace(params NewWorkspaceParams) Workspace {
	return &realWorkspace{
		fs:               params.FS,
		git:              params.Git,
		config:           params.Config,
		statusManager:    params.StatusManager,
		logger:           params.Logger,
		prompt:           params.Prompt,
		worktreeProvider: params.WorktreeProvider,
	}
}

// SetLogger sets the logger for this workspace instance.
func (w *realWorkspace) SetLogger(logger logger.Logger) {
	w.logger = logger
}

// CreateWorktreeOpts contains optional parameters for worktree creation.
type CreateWorktreeOpts struct {
	IDEName string
}

// LoadWorktreeOpts contains optional parameters for LoadWorktree.
type LoadWorktreeOpts struct {
	IDEName string
}

// WorktreeInfo represents information about a worktree.
type WorktreeInfo struct {
	Path   string
	Branch string
	Status string
}

// Load handles the complete workspace loading workflow.
// It detects workspace files, handles user selection if multiple files are found,
// and loads the workspace configuration for display.
func (w *realWorkspace) Load(force bool) error {
	// If already loaded, just parse and display the configuration
	if w.OriginalFile != "" {
		workspaceConfig, err := w.ParseFile(w.OriginalFile)
		if err != nil {
			return fmt.Errorf("failed to parse workspace file: %w", err)
		}

		w.logger.Logf("Workspace mode detected")

		workspaceName := w.GetName(workspaceConfig, w.OriginalFile)
		w.logger.Logf("Found workspace: %s", workspaceName)

		w.logger.Logf("Workspace configuration:")
		w.logger.Logf("  Folders: %d", len(workspaceConfig.Folders))
		for _, folder := range workspaceConfig.Folders {
			w.logger.Logf("    - %s: %s", folder.Name, folder.Path)
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
		selectedFile, err := w.HandleMultipleFiles(workspaceFiles, force)
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

	w.logger.Logf("Workspace mode detected")

	workspaceName := w.GetName(workspaceConfig, w.OriginalFile)
	w.logger.Logf("Found workspace: %s", workspaceName)

	w.logger.Logf("Workspace configuration:")
	w.logger.Logf("  Folders: %d", len(workspaceConfig.Folders))
	for _, folder := range workspaceConfig.Folders {
		w.logger.Logf("    - %s: %s", folder.Name, folder.Path)
	}

	return nil
}

// Validate validates all repositories in a workspace.
func (w *realWorkspace) Validate() error {
	w.logger.Logf("Validating workspace: %s", w.OriginalFile)

	// Use the new workspace validation logic that ensures repositories are in status
	// and have default branch worktrees
	return w.ValidateWorkspaceReferences()
}

// ListWorktrees lists worktrees for workspace mode.
func (w *realWorkspace) ListWorktrees(force bool) ([]status.WorktreeInfo, error) {
	w.logger.Logf("Listing worktrees for workspace mode")

	// Load workspace configuration (only if not already loaded)
	if w.OriginalFile == "" {
		if err := w.Load(force); err != nil {
			return nil, fmt.Errorf("failed to load workspace: %w", err)
		}
	}

	// Get workspace path
	workspacePath, err := w.getWorkspacePath()
	if err != nil {
		return nil, err
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

// CreateWorktree creates worktrees for all repositories in the workspace.
func (w *realWorkspace) CreateWorktree(branch string, force bool, opts ...CreateWorktreeOpts) (string, error) {
	w.logger.Logf("Creating worktrees for branch: %s", branch)

	// 1. Load and validate workspace configuration (only if not already loaded)
	if w.OriginalFile == "" {
		if err := w.Load(force); err != nil {
			return "", fmt.Errorf("failed to load workspace: %w", err)
		}
	}

	// 2. Validate all repositories in workspace
	if err := w.Validate(); err != nil {
		return "", fmt.Errorf("failed to validate workspace: %w", err)
	}

	// 3. Pre-validate worktree creation for all repositories
	if err := w.validateWorkspaceForWorktreeCreation(branch); err != nil {
		return "", fmt.Errorf("failed to validate workspace for worktree creation: %w", err)
	}

	// 4. Create worktrees for all repositories
	var workspaceOpts *CreateWorktreeOpts
	if len(opts) > 0 {
		workspaceOpts = &opts[0]
	}
	if err := w.createWorktreesForWorkspace(branch, workspaceOpts); err != nil {
		return "", fmt.Errorf("failed to create worktrees: %w", err)
	}

	// 5. Calculate and return the worktree path
	worktreePath := filepath.Join(
		w.config.BasePath,
		"workspaces",
		fmt.Sprintf("workspace-%s", branch),
	)

	w.logger.Logf("Workspace worktree creation completed successfully")
	return worktreePath, nil
}

// DeleteWorktree deletes worktrees for the workspace with the specified branch.
func (w *realWorkspace) DeleteWorktree(branch string, force bool) error {
	w.logger.Logf("Deleting worktrees for branch: %s", branch)

	// Load workspace configuration (only if not already loaded)
	if w.OriginalFile == "" {
		if err := w.Load(force); err != nil {
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
		w.logger.Logf("Warning: failed to remove worktree workspace file: %v", err)
	}

	// Remove worktree entries from status file
	if err := w.removeWorktreeStatusEntries(workspaceWorktrees, force); err != nil {
		return err
	}

	w.logger.Logf("Workspace worktree deletion completed successfully")
	return nil
}
