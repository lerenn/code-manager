package ide

import (
	"fmt"

	"github.com/lerenn/code-manager/pkg/fs"
)

const (
	// VSCodiumName is the name identifier for the VSCodium IDE.
	VSCodiumName = "vscodium"
	// VSCodiumCommand is the command to open VSCodium.
	VSCodiumCommand = "codium"
)

// VSCodium represents the VSCodium IDE implementation.
type VSCodium struct {
	fs fs.FS
}

// NewVSCodium creates a new VSCodium IDE instance.
func NewVSCodium(fs fs.FS) *VSCodium {
	return &VSCodium{
		fs: fs,
	}
}

// Name returns the name of the IDE.
func (v *VSCodium) Name() string {
	return VSCodiumName
}

// IsInstalled checks if VSCodium is installed on the system.
func (v *VSCodium) IsInstalled() bool {
	_, err := v.fs.Which(VSCodiumCommand)
	return err == nil
}

// OpenRepository opens VSCodium with the specified repository path.
func (v *VSCodium) OpenRepository(path string) error {
	// Execute codium command with the repository path
	if err := v.fs.ExecuteCommand(VSCodiumCommand, path); err != nil {
		return fmt.Errorf("%w: %s", ErrIDEExecutionFailed, VSCodiumCommand)
	}
	return nil
}
