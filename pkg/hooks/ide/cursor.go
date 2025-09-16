// Package ide provides IDE opening functionality through hooks.
package ide

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/fs"
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
	// Ensure the path is absolute and has a trailing slash for new window
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Add trailing slash to ensure Cursor opens a new window
	if absPath[len(absPath)-1] != '/' {
		absPath += "/"
	}

	// Execute cursor command with the absolute path
	if err := c.fs.ExecuteCommand(CursorCommand, absPath); err != nil {
		return fmt.Errorf("%w: %s", ErrIDEExecutionFailed, CursorCommand)
	}
	return nil
}
