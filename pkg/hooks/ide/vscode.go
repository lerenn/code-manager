package ide

import (
	"fmt"
	"path/filepath"

	"github.com/lerenn/code-manager/pkg/fs"
)

const (
	// VSCodeName is the name identifier for the VS Code IDE.
	VSCodeName = "vscode"
	// VSCodeCommand is the command to open VS Code.
	VSCodeCommand = "code"
)

// VSCode represents the VS Code IDE implementation.
type VSCode struct {
	fs fs.FS
}

// NewVSCode creates a new VS Code IDE instance.
func NewVSCode(fs fs.FS) *VSCode {
	return &VSCode{
		fs: fs,
	}
}

// Name returns the name of the IDE.
func (v *VSCode) Name() string {
	return VSCodeName
}

// IsInstalled checks if VS Code is installed on the system.
func (v *VSCode) IsInstalled() bool {
	_, err := v.fs.Which(VSCodeCommand)
	return err == nil
}

// OpenRepository opens VS Code with the specified repository path.
func (v *VSCode) OpenRepository(path string) error {
	// Ensure the path is absolute and has a trailing slash for new window
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Add trailing slash to ensure VS Code opens a new window
	if absPath[len(absPath)-1] != '/' {
		absPath += "/"
	}

	// Execute code command with the absolute path
	if err := v.fs.ExecuteCommand(VSCodeCommand, absPath); err != nil {
		return fmt.Errorf("%w: %s", ErrIDEExecutionFailed, VSCodeCommand)
	}
	return nil
}
