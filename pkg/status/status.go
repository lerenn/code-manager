package status

import (
	"fmt"

	"github.com/lerenn/cgwt/pkg/config"
	"github.com/lerenn/cgwt/pkg/fs"
	"gopkg.in/yaml.v3"
)

//go:generate go run go.uber.org/mock/mockgen@v0.5.2 -source=status.go -destination=mockstatus.gen.go -package=status

// Status represents the status.yaml file structure.
type Status struct {
	Repositories []Repository `yaml:"repositories"`
}

// Repository represents a repository entry in the status file.
type Repository struct {
	Name      string `yaml:"name"`
	Branch    string `yaml:"branch"`
	Path      string `yaml:"path"`
	Workspace string `yaml:"workspace,omitempty"`
}

// Manager interface provides status file management functionality.
type Manager interface {
	// AddWorktree adds a worktree entry to the status file.
	AddWorktree(repoName, branch, worktreePath, workspacePath string) error
	// RemoveWorktree removes a worktree entry from the status file.
	RemoveWorktree(repoName, branch string) error
	// GetWorktree retrieves the status of a specific worktree.
	GetWorktree(repoName, branch string) (*Repository, error)
	// ListAllWorktrees lists all tracked worktrees.
	ListAllWorktrees() ([]Repository, error)
}

type realManager struct {
	fs     fs.FS
	config *config.Config
}

// NewManager creates a new Status Manager instance.
func NewManager(fs fs.FS, config *config.Config) Manager {
	return &realManager{
		fs:     fs,
		config: config,
	}
}

// AddWorktree adds a worktree entry to the status file.
func (s *realManager) AddWorktree(repoName, branch, worktreePath, workspacePath string) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Check for duplicate entry
	for _, repo := range status.Repositories {
		if repo.Name == repoName && repo.Branch == branch {
			return fmt.Errorf("worktree already exists for repository %s branch %s", repoName, branch)
		}
	}

	// Create new repository entry
	newRepo := Repository{
		Name:      repoName,
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
func (s *realManager) RemoveWorktree(repoName, branch string) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Find and remove the repository entry
	found := false
	var newRepositories []Repository
	for _, repo := range status.Repositories {
		if repo.Name == repoName && repo.Branch == branch {
			found = true
			continue // Skip this entry
		}
		newRepositories = append(newRepositories, repo)
	}

	if !found {
		return fmt.Errorf("worktree not found for repository %s branch %s", repoName, branch)
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
func (s *realManager) GetWorktree(repoName, branch string) (*Repository, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	// Find the repository entry
	for _, repo := range status.Repositories {
		if repo.Name == repoName && repo.Branch == branch {
			return &repo, nil
		}
	}

	return nil, fmt.Errorf("worktree not found for repository %s branch %s", repoName, branch)
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

// getStatusFilePath returns the status file path from configuration.
func (s *realManager) getStatusFilePath() (string, error) {
	if s.config == nil {
		return "", fmt.Errorf("configuration is not initialized")
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

	return nil
}
