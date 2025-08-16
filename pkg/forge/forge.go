package forge

import (
	"fmt"

	"github.com/lerenn/wtm/pkg/logger"
)

//go:generate mockgen -source=forge.go -destination=mockforge.gen.go -package=forge

// IssueInfo represents information about a forge issue.
type IssueInfo struct {
	Number      int
	Title       string
	Description string
	State       string
	URL         string
	Repository  string
	Owner       string
}

// IssueReference represents a parsed issue reference.
type IssueReference struct {
	Owner       string
	Repository  string
	IssueNumber int
	URL         string
}

// Forge interface defines the methods that all forge implementations must provide.
type Forge interface {
	// Name returns the name of the forge
	Name() string

	// GetIssueInfo fetches issue information from the forge
	GetIssueInfo(issueRef string) (*IssueInfo, error)

	// ValidateForgeRepository validates that repository has supported forge remote origin
	ValidateForgeRepository(repoPath string) error

	// ParseIssueReference parses various issue reference formats
	ParseIssueReference(issueRef string) (*IssueReference, error)

	// GenerateBranchName generates branch name from issue information
	GenerateBranchName(issueInfo *IssueInfo) string
}

// ManagerInterface defines the interface for forge management.
type ManagerInterface interface {
	// GetForge returns the forge implementation for the given name
	GetForge(name string) (Forge, error)
	// GetForgeForRepository returns the appropriate forge for the given repository
	GetForgeForRepository(repoPath string) (Forge, error)
}

// Manager manages forge implementations and provides a unified interface.
type Manager struct {
	forges map[string]Forge
	logger logger.Logger
}

// NewManager creates a new forge manager with registered forge implementations.
func NewManager(logger logger.Logger) *Manager {
	m := &Manager{
		forges: make(map[string]Forge),
		logger: logger,
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

// GetForge returns the forge implementation for the given name.
func (m *Manager) GetForge(name string) (Forge, error) {
	forge, exists := m.forges[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedForge, name)
	}
	return forge, nil
}

// GetForgeForRepository returns the appropriate forge for the given repository.
func (m *Manager) GetForgeForRepository(repoPath string) (Forge, error) {
	// Try each forge to see which one can validate the repository
	for _, forge := range m.forges {
		if err := forge.ValidateForgeRepository(repoPath); err == nil {
			return forge, nil
		}
	}
	return nil, fmt.Errorf("%w: no supported forge found for repository", ErrUnsupportedForge)
}
