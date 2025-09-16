package ide

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/fs"
)

const (
	// DummyName is the name identifier for the Dummy IDE.
	DummyName = "dummy"
)

// Dummy represents a dummy IDE implementation for testing.
type Dummy struct {
	fs fs.FS
}

// NewDummy creates a new Dummy IDE instance.
func NewDummy(fs fs.FS) *Dummy {
	return &Dummy{
		fs: fs,
	}
}

// Name returns the name of the IDE.
func (d *Dummy) Name() string {
	return DummyName
}

// IsInstalled always returns true for the dummy IDE.
func (d *Dummy) IsInstalled() bool {
	return true
}

// OpenRepository does nothing but returns success.
func (d *Dummy) OpenRepository(path string) error {
	// Ensure the path is absolute and has a trailing slash for consistency
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Add trailing slash for consistency with other IDEs
	if absPath[len(absPath)-1] != '/' {
		absPath += "/"
	}

	// Dummy IDE prints the path for testing purposes
	fmt.Println("DUMMY_IDE_PATH:", absPath)
	// Dummy IDE does nothing, just returns success
	return nil
}
