package status

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"gopkg.in/yaml.v3"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=status.go -destination=mockstatus.gen.go -package=status

// Status represents the status.yaml file structure.
type Status struct {
	Repositories []Repository `yaml:"repositories"`
}

// Repository represents a repository entry in the status file.
type Repository struct {
	URL       string `yaml:"url"`                 // Repository URL (e.g., "github.com/lerenn/wtm")
	Branch    string `yaml:"branch"`              // Branch name
	Path      string `yaml:"path"`                // Original repository path (not worktree path)
	Workspace string `yaml:"workspace,omitempty"` // Workspace path (if applicable)
	Remote    string `yaml:"remote,omitempty"`    // Remote name (e.g., "origin", "justenstall")
}

// Manager interface provides status file management functionality.
type Manager interface {
	// AddWorktree adds a worktree entry to the status file.
	AddWorktree(repoURL, branch, worktreePath, workspacePath string) error
	// RemoveWorktree removes a worktree entry from the status file.
	RemoveWorktree(repoURL, branch string) error
	// GetWorktree retrieves the status of a specific worktree.
	GetWorktree(repoURL, branch string) (*Repository, error)
	// ListAllWorktrees lists all tracked worktrees.
	ListAllWorktrees() ([]Repository, error)
	// GetWorkspaceWorktrees returns all worktrees for a specific workspace and branch.
	GetWorkspaceWorktrees(workspacePath, branchName string) ([]Repository, error)
	// GetWorkspaceBranches returns all branch names for a specific workspace.
	GetWorkspaceBranches(workspacePath string) ([]string, error)
}

type realManager struct {
	fs         fs.FS
	config     *config.Config
	workspaces map[string]map[string][]Repository // workspace -> branch -> repositories
}

// NewManager creates a new Status Manager instance.
func NewManager(fs fs.FS, config *config.Config) Manager {
	manager := &realManager{
		fs:         fs,
		config:     config,
		workspaces: make(map[string]map[string][]Repository),
	}

	// Initialize workspaces map
	manager.initializeWorkspacesMap()

	return manager
}

// initializeWorkspacesMap loads the status and computes the workspaces map.
func (s *realManager) initializeWorkspacesMap() {
	status, err := s.loadStatus()
	if err != nil {
		// If we can't load status, start with empty map
		s.workspaces = make(map[string]map[string][]Repository)
		return
	}

	s.computeWorkspacesMap(status.Repositories)
}

// AddWorktree adds a worktree entry to the status file.
func (s *realManager) AddWorktree(repoURL, branch, worktreePath, workspacePath string) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Check for duplicate entry
	for _, repo := range status.Repositories {
		if repo.URL == repoURL && repo.Branch == branch {
			return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeAlreadyExists, repoURL, branch)
		}
	}

	// Create new repository entry
	newRepo := Repository{
		URL:       repoURL,
		Branch:    branch,
		Path:      worktreePath,
		Workspace: workspacePath,
	}

	// Add to repositories list
	status.Repositories = append(status.Repositories, newRepo)

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}

// RemoveWorktree removes a worktree entry from the status file.
func (s *realManager) RemoveWorktree(repoURL, branch string) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Find and remove the repository entry
	found := false
	var newRepositories []Repository
	for _, repo := range status.Repositories {
		if repo.URL == repoURL && repo.Branch == branch {
			found = true
			continue // Skip this entry
		}
		newRepositories = append(newRepositories, repo)
	}

	if !found {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotFound, repoURL, branch)
	}

	// Update repositories list
	status.Repositories = newRepositories

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}

// GetWorktree retrieves the status of a specific worktree.
func (s *realManager) GetWorktree(repoURL, branch string) (*Repository, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	// Find the repository entry
	for _, repo := range status.Repositories {
		if repo.URL == repoURL && repo.Branch == branch {
			return &repo, nil
		}
	}

	return nil, fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotFound, repoURL, branch)
}

// ListAllWorktrees lists all tracked worktrees.
func (s *realManager) ListAllWorktrees() ([]Repository, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	return status.Repositories, nil
}

// computeWorkspacesMap computes the workspaces map from the repositories list.
func (s *realManager) computeWorkspacesMap(repositories []Repository) {
	s.workspaces = make(map[string]map[string][]Repository)

	for _, repo := range repositories {
		if repo.Workspace == "" {
			continue // Skip non-workspace repositories
		}

		// Get workspace name from the workspace path
		workspaceName := s.getWorkspaceNameFromPath(repo.Workspace)

		// Initialize workspace map if it doesn't exist
		if s.workspaces[workspaceName] == nil {
			s.workspaces[workspaceName] = make(map[string][]Repository)
		}

		// Add repository to the appropriate branch
		s.workspaces[workspaceName][repo.Branch] = append(s.workspaces[workspaceName][repo.Branch], repo)
	}
}

// getWorkspaceNameFromPath extracts the workspace name from the workspace file path.
func (s *realManager) getWorkspaceNameFromPath(workspacePath string) string {
	// Extract filename without extension
	// This is a simple implementation - in practice, you might want to parse the workspace file
	// to get the actual workspace name from the JSON content
	// For now, we'll use the filename without .code-workspace extension
	// This matches the logic in the workspace.go getName method
	return strings.TrimSuffix(filepath.Base(workspacePath), ".code-workspace")
}

// GetWorkspaceWorktrees returns all worktrees for a specific workspace and branch.
func (s *realManager) GetWorkspaceWorktrees(workspacePath, branchName string) ([]Repository, error) {
	workspaceName := s.getWorkspaceNameFromPath(workspacePath)
	if s.workspaces[workspaceName] == nil {
		return []Repository{}, nil
	}

	return s.workspaces[workspaceName][branchName], nil
}

// GetWorkspaceBranches returns all branch names for a specific workspace.
func (s *realManager) GetWorkspaceBranches(workspacePath string) ([]string, error) {
	workspaceName := s.getWorkspaceNameFromPath(workspacePath)
	if s.workspaces[workspaceName] == nil {
		return []string{}, nil
	}

	var branches []string
	for branch := range s.workspaces[workspaceName] {
		branches = append(branches, branch)
	}

	return branches, nil
}

// getStatusFilePath returns the status file path from configuration.
func (s *realManager) getStatusFilePath() (string, error) {
	if s.config == nil {
		return "", ErrConfigurationNotInitialized
	}

	if s.config.StatusFile == "" {
		return "", fmt.Errorf("status file path is not configured")
	}

	return s.config.StatusFile, nil
}

// loadStatus loads the status from the status file.
func (s *realManager) loadStatus() (*Status, error) {
	statusPath, err := s.getStatusFilePath()
	if err != nil {
		return nil, fmt.Errorf("failed to get status file path: %w", err)
	}

	// Check if status file exists
	exists, err := s.fs.Exists(statusPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check status file existence: %w", err)
	}

	if !exists {
		// Create initial status file
		initialStatus := &Status{
			Repositories: []Repository{},
		}
		if err := s.saveStatus(initialStatus); err != nil {
			return nil, fmt.Errorf("failed to create initial status file: %w", err)
		}
		return initialStatus, nil
	}

	// Read status file
	data, err := s.fs.ReadFile(statusPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read status file: %w", err)
	}

	// Parse YAML
	var status Status
	if err := yaml.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status file: %w", err)
	}

	return &status, nil
}

// saveStatus saves the status to the status file atomically.
func (s *realManager) saveStatus(status *Status) error {
	statusPath, err := s.getStatusFilePath()
	if err != nil {
		return fmt.Errorf("failed to get status file path: %w", err)
	}

	// Acquire file lock
	unlock, err := s.fs.FileLock(statusPath)
	if err != nil {
		return fmt.Errorf("failed to acquire file lock: %w", err)
	}
	defer unlock()

	// Marshal status to YAML
	data, err := yaml.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// Write status file atomically
	if err := s.fs.WriteFileAtomic(statusPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write status file: %w", err)
	}

	// Update workspaces map after successful save
	s.computeWorkspacesMap(status.Repositories)

	return nil
}
