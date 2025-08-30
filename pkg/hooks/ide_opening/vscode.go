package ide_opening

import (
	"fmt"

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
	// Execute code command with the repository path
	if err := v.fs.ExecuteCommand(VSCodeCommand, path); err != nil {
		return fmt.Errorf("%w: %s", ErrIDEExecutionFailed, VSCodeCommand)
	}
	return nil
}
