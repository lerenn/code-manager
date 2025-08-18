// Package ide provides interfaces and implementations for interacting with various IDEs.
package ide

import (
	"fmt"

	"github.com/lerenn/cm/pkg/fs"
)

const (
	// CursorName is the name identifier for the Cursor IDE.
	CursorName = "cursor"
	// CursorCommand is the command to open Cursor.
	CursorCommand = "cursor"
)

// Cursor represents the Cursor IDE implementation.
type Cursor struct {
	fs fs.FS
}

// NewCursor creates a new Cursor IDE instance.
func NewCursor(fs fs.FS) *Cursor {
	return &Cursor{
		fs: fs,
	}
}

// Name returns the name of the IDE.
func (c *Cursor) Name() string {
	return CursorName
}

// IsInstalled checks if Cursor is installed on the system.
func (c *Cursor) IsInstalled() bool {
	_, err := c.fs.Which(CursorCommand)
	return err == nil
}

// OpenRepository opens Cursor with the specified repository path.
func (c *Cursor) OpenRepository(path string) error {
	// Execute cursor command with the repository path
	if err := c.fs.ExecuteCommand(CursorCommand, path); err != nil {
		return fmt.Errorf("%w: %s", ErrIDEExecutionFailed, CursorCommand)
	}
	return nil
}
