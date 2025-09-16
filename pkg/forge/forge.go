package forge

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/issue"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
)

//go:generate mockgen -source=forge.go -destination=mocks/forge.gen.go -package=mocks

// Forge interface defines the methods that all forge implementations must provide.
type Forge interface {
	// Name returns the name of the forge
	Name() string

	// GetIssueInfo fetches issue information from the forge
	GetIssueInfo(issueRef string) (*issue.Info, error)

	// ValidateForgeRepository validates that repository has supported forge remote origin
	ValidateForgeRepository(repoPath string) error

	// ParseIssueReference parses various issue reference formats
	ParseIssueReference(issueRef string) (*issue.Reference, error)

	// GenerateBranchName generates branch name from issue information
	GenerateBranchName(issueInfo *issue.Info) string
}

// ManagerInterface defines the interface for forge management.
type ManagerInterface interface {
	// GetForgeForRepository returns the appropriate forge for the given repository
	GetForgeForRepository(repoName string) (Forge, error)
}

// Manager manages forge implementations and provides a unified interface.
type Manager struct {
	forges        map[string]Forge
	logger        logger.Logger
	statusManager status.Manager
}

// NewManager creates a new forge manager with registered forge implementations.
func NewManager(logger logger.Logger, statusManager status.Manager) *Manager {
	m := &Manager{
		forges:        make(map[string]Forge),
		logger:        logger,
		statusManager: statusManager,
	}

	// Register forge implementations
	m.registerForges()

	return m
}

// registerForges registers all available forge implementations.
func (m *Manager) registerForges() {
	// Register GitHub forge
	github := NewGitHub()
	m.forges[github.Name()] = github
}

// GetForgeForRepository returns the appropriate forge for the given repository.
func (m *Manager) GetForgeForRepository(repoName string) (Forge, error) {
	// Resolve the repository name to a path
	repoPath, err := m.resolveRepository(repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repository '%s': %w", repoName, err)
	}

	// Try each forge to see which one can validate the repository
	for _, forge := range m.forges {
		if err := forge.ValidateForgeRepository(repoPath); err == nil {
			return forge, nil
		}
	}
	return nil, fmt.Errorf("%w: no supported forge found for repository", ErrUnsupportedForge)
}

// resolveRepository resolves a repository name to a path, checking status file first.
func (m *Manager) resolveRepository(repoName string) (string, error) {
	// If empty, use current directory
	if repoName == "" {
		return ".", nil
	}

	// First, check if it's a repository name from status.yaml
	if existingRepo, err := m.statusManager.GetRepository(repoName); err == nil && existingRepo != nil {
		m.logger.Logf("Resolved repository '%s' from status.yaml: %s", repoName, existingRepo.Path)
		return existingRepo.Path, nil
	}

	// Check if it's an absolute path
	if filepath.IsAbs(repoName) {
		return repoName, nil
	}

	// Resolve relative path from current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	resolvedPath := filepath.Join(currentDir, repoName)
	return resolvedPath, nil
}
