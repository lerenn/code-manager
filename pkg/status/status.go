package status

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/issue"
	"gopkg.in/yaml.v3"
)

//go:generate mockgen -source=status.go -destination=mockstatus.gen.go -package=status

// Status represents the status.yaml file structure.
type Status struct {
	Repositories map[string]Repository `yaml:"repositories"`
	Workspaces   map[string]Workspace  `yaml:"workspaces"`
}

// Repository represents a repository entry in the status file.
type Repository struct {
	Path      string                  `yaml:"path"`
	Remotes   map[string]Remote       `yaml:"remotes"`
	Worktrees map[string]WorktreeInfo `yaml:"worktrees"`
}

// Remote represents a remote configuration for a repository.
type Remote struct {
	DefaultBranch string `yaml:"default_branch"`
}

// Workspace represents a workspace entry in the status file.
type Workspace struct {
	Worktree     string   `yaml:"worktree"`
	Repositories []string `yaml:"repositories"`
}

// WorktreeInfo represents worktree information.
type WorktreeInfo struct {
	Remote string      `yaml:"remote"`
	Branch string      `yaml:"branch"`
	Issue  *issue.Info `yaml:"issue,omitempty"`
}

// Manager interface provides status file management functionality.
type Manager interface {
	// AddWorktree adds a worktree entry to the status file.
	AddWorktree(params AddWorktreeParams) error
	// RemoveWorktree removes a worktree entry from the status file.
	RemoveWorktree(repoURL, branch string) error
	// GetWorktree retrieves the status of a specific worktree.
	GetWorktree(repoURL, branch string) (*WorktreeInfo, error)
	// ListAllWorktrees lists all tracked worktrees.
	ListAllWorktrees() ([]WorktreeInfo, error)
	// GetWorkspaceWorktrees returns all worktrees for a specific workspace and branch.
	GetWorkspaceWorktrees(workspacePath, branchName string) ([]WorktreeInfo, error)
	// GetWorkspaceBranches returns all branch names for a specific workspace.
	GetWorkspaceBranches(workspacePath string) ([]string, error)
	// CreateInitialStatus creates the initial status file structure.
	CreateInitialStatus() error
	// AddRepository adds a repository entry to the status file.
	AddRepository(repoURL string, params AddRepositoryParams) error
	// GetRepository retrieves a repository entry from the status file.
	GetRepository(repoURL string) (*Repository, error)
	// ListRepositories lists all repositories in the status file.
	ListRepositories() (map[string]Repository, error)
	// AddWorkspace adds a workspace entry to the status file.
	AddWorkspace(workspacePath string, params AddWorkspaceParams) error
	// GetWorkspace retrieves a workspace entry from the status file.
	GetWorkspace(workspacePath string) (*Workspace, error)
	// ListWorkspaces lists all workspaces in the status file.
	ListWorkspaces() (map[string]Workspace, error)
}

type realManager struct {
	fs         fs.FS
	config     *config.Config
	workspaces map[string]map[string][]WorktreeInfo // workspace -> branch -> worktrees
}

// NewManager creates a new Status Manager instance.
func NewManager(fs fs.FS, config *config.Config) Manager {
	manager := &realManager{
		fs:         fs,
		config:     config,
		workspaces: make(map[string]map[string][]WorktreeInfo),
	}

	// Initialize workspaces map
	manager.initializeWorkspacesMap()

	return manager
}

// CreateInitialStatus creates the initial status file structure.
func (s *realManager) CreateInitialStatus() error {
	initialStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	return s.saveStatus(initialStatus)
}

// initializeWorkspacesMap loads the status and computes the workspaces map.
func (s *realManager) initializeWorkspacesMap() {
	status, err := s.loadStatus()
	if err != nil {
		// If we can't load status, start with empty map
		s.workspaces = make(map[string]map[string][]WorktreeInfo)
		return
	}

	s.computeWorkspacesMap(status.Workspaces)
}

// AddWorktreeParams contains parameters for AddWorktree.
type AddWorktreeParams struct {
	RepoURL       string
	Branch        string
	WorktreePath  string
	WorkspacePath string
	IssueInfo     *issue.Info
	Remote        string
}

// AddRepositoryParams contains parameters for AddRepository.
type AddRepositoryParams struct {
	Path    string
	Remotes map[string]Remote
}

// AddWorkspaceParams contains parameters for AddWorkspace.
type AddWorkspaceParams struct {
	Worktree     string
	Repositories []string
}

// AddWorktree adds a worktree entry to the status file.
func (s *realManager) AddWorktree(params AddWorktreeParams) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Ensure repository exists
	if _, exists := status.Repositories[params.RepoURL]; !exists {
		return fmt.Errorf("%w: %s", ErrRepositoryNotFound, params.RepoURL)
	}

	// Check for duplicate worktree entry
	worktreeKey := fmt.Sprintf("%s:%s", params.Remote, params.Branch)
	if _, exists := status.Repositories[params.RepoURL].Worktrees[worktreeKey]; exists {
		return fmt.Errorf("%w for repository %s worktree %s", ErrWorktreeAlreadyExists, params.RepoURL, worktreeKey)
	}

	// Create new worktree entry
	worktreeInfo := WorktreeInfo{
		Remote: params.Remote,
		Branch: params.Branch,
		Issue:  params.IssueInfo,
	}

	// Add to repository's worktrees
	repo := status.Repositories[params.RepoURL]
	if repo.Worktrees == nil {
		repo.Worktrees = make(map[string]WorktreeInfo)
	}
	repo.Worktrees[worktreeKey] = worktreeInfo
	status.Repositories[params.RepoURL] = repo

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

	// Check if repository exists
	repo, exists := status.Repositories[repoURL]
	if !exists {
		return fmt.Errorf("%w: %s", ErrRepositoryNotFound, repoURL)
	}

	// Find and remove the worktree entry
	found := false
	for worktreeKey, worktree := range repo.Worktrees {
		if worktree.Branch == branch {
			delete(repo.Worktrees, worktreeKey)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotFound, repoURL, branch)
	}

	// Update repository
	status.Repositories[repoURL] = repo

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}

// GetWorktree retrieves the status of a specific worktree.
func (s *realManager) GetWorktree(repoURL, branch string) (*WorktreeInfo, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	// Check if repository exists
	repo, exists := status.Repositories[repoURL]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrRepositoryNotFound, repoURL)
	}

	// Find the worktree entry
	for _, worktree := range repo.Worktrees {
		if worktree.Branch == branch {
			return &worktree, nil
		}
	}

	return nil, fmt.Errorf("%w for repository %s branch %s", ErrWorktreeNotFound, repoURL, branch)
}

// ListAllWorktrees lists all tracked worktrees.
func (s *realManager) ListAllWorktrees() ([]WorktreeInfo, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	var worktrees []WorktreeInfo
	for _, repo := range status.Repositories {
		for _, worktree := range repo.Worktrees {
			worktrees = append(worktrees, worktree)
		}
	}

	return worktrees, nil
}

// AddRepository adds a repository entry to the status file.
func (s *realManager) AddRepository(repoURL string, params AddRepositoryParams) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Check for duplicate repository
	if _, exists := status.Repositories[repoURL]; exists {
		return fmt.Errorf("%w: %s", ErrRepositoryAlreadyExists, repoURL)
	}

	// Create new repository entry
	repo := Repository{
		Path:      params.Path,
		Remotes:   params.Remotes,
		Worktrees: make(map[string]WorktreeInfo),
	}

	// Add to repositories map
	status.Repositories[repoURL] = repo

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}

// GetRepository retrieves a repository entry from the status file.
func (s *realManager) GetRepository(repoURL string) (*Repository, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	// Check if repository exists
	repo, exists := status.Repositories[repoURL]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrRepositoryNotFound, repoURL)
	}

	return &repo, nil
}

// ListRepositories lists all repositories in the status file.
func (s *realManager) ListRepositories() (map[string]Repository, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	return status.Repositories, nil
}

// AddWorkspace adds a workspace entry to the status file.
func (s *realManager) AddWorkspace(workspacePath string, params AddWorkspaceParams) error {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return fmt.Errorf("failed to load status: %w", err)
	}

	// Check for duplicate workspace
	if _, exists := status.Workspaces[workspacePath]; exists {
		return fmt.Errorf("%w: %s", ErrWorkspaceAlreadyExists, workspacePath)
	}

	// Create new workspace entry
	workspace := Workspace(params)

	// Add to workspaces map
	status.Workspaces[workspacePath] = workspace

	// Save updated status
	if err := s.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}

// GetWorkspace retrieves a workspace entry from the status file.
func (s *realManager) GetWorkspace(workspacePath string) (*Workspace, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	// Check if workspace exists
	workspace, exists := status.Workspaces[workspacePath]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrWorkspaceNotFound, workspacePath)
	}

	return &workspace, nil
}

// ListWorkspaces lists all workspaces in the status file.
func (s *realManager) ListWorkspaces() (map[string]Workspace, error) {
	// Load current status
	status, err := s.loadStatus()
	if err != nil {
		return nil, fmt.Errorf("failed to load status: %w", err)
	}

	return status.Workspaces, nil
}

// computeWorkspacesMap computes the workspaces map from the workspaces list.
func (s *realManager) computeWorkspacesMap(workspaces map[string]Workspace) {
	s.workspaces = make(map[string]map[string][]WorktreeInfo)

	for workspacePath := range workspaces {
		// Get workspace name from the workspace path
		workspaceName := s.getWorkspaceNameFromPath(workspacePath)

		// Initialize workspace map if it doesn't exist
		if s.workspaces[workspaceName] == nil {
			s.workspaces[workspaceName] = make(map[string][]WorktreeInfo)
		}

		// For now, we'll store the worktree reference
		// In the future, this could be enhanced to store actual worktree info
		// based on the worktree reference in the workspace
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
func (s *realManager) GetWorkspaceWorktrees(workspacePath, branchName string) ([]WorktreeInfo, error) {
	workspaceName := s.getWorkspaceNameFromPath(workspacePath)
	if s.workspaces[workspaceName] == nil {
		return []WorktreeInfo{}, nil
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
			Repositories: make(map[string]Repository),
			Workspaces:   make(map[string]Workspace),
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
	s.computeWorkspacesMap(status.Workspaces)

	return nil
}
