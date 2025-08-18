package ide

import (
	"fmt"

	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/logger"
)

//go:generate mockgen -source=ide.go -destination=mockide.gen.go -package=ide

// IDE interface defines the methods that all IDE implementations must provide.
type IDE interface {
	// Name returns the name of the IDE
	Name() string

	// IsInstalled checks if the IDE is installed on the system
	IsInstalled() bool

	// OpenRepository opens the IDE with the specified repository path
	OpenRepository(path string) error
}

// ManagerInterface defines the interface for IDE management.
type ManagerInterface interface {
	// GetIDE returns the IDE implementation for the given name
	GetIDE(name string) (IDE, error)
	// OpenIDE opens the specified IDE with the given path
	OpenIDE(name, path string, verbose bool) error
}

// Manager manages IDE implementations and provides a unified interface.
type Manager struct {
	ides   map[string]IDE
	logger logger.Logger
}

// NewManager creates a new IDE manager with registered IDE implementations.
func NewManager(fs fs.FS, logger logger.Logger) *Manager {
	m := &Manager{
		ides:   make(map[string]IDE),
		logger: logger,
	}

	// Register IDE implementations
	m.registerIDEs(fs)

	return m
}

// registerIDEs registers all available IDE implementations.
func (m *Manager) registerIDEs(fs fs.FS) {
	// Register Cursor IDE
	cursor := NewCursor(fs)
	m.ides[cursor.Name()] = cursor

	// Register Dummy IDE for testing
	dummy := NewDummy(fs)
	m.ides[dummy.Name()] = dummy
}

// GetIDE returns the IDE implementation for the given name.
func (m *Manager) GetIDE(name string) (IDE, error) {
	ide, exists := m.ides[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedIDE, name)
	}
	return ide, nil
}

// OpenIDE opens the specified IDE with the given path.
func (m *Manager) OpenIDE(name, path string, verbose bool) error {
	ide, err := m.GetIDE(name)
	if err != nil {
		return err
	}

	// Check if IDE is installed
	if !ide.IsInstalled() {
		return fmt.Errorf("%w: %s", ErrIDENotInstalled, name)
	}

	// Log the path being opened if verbose is enabled
	if verbose {
		m.logger.Logf("Opening %s with %s at path: %s", name, name, path)
	}

	// Open the repository in the IDE
	if err := ide.OpenRepository(path); err != nil {
		m.logger.Logf("Failed to open %s: %v", name, err)
		return fmt.Errorf("%w: %s", err, name)
	}

	// Log success if verbose is enabled
	if verbose {
		m.logger.Logf("Successfully opened %s with %s", path, name)
	}

	return nil
}
