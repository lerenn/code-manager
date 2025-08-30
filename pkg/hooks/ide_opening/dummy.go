package ide_opening

import (
	"fmt"

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
	// Dummy IDE prints the path for testing purposes
	fmt.Println("DUMMY_IDE_PATH:", path)
	// Dummy IDE does nothing, just returns success
	return nil
}
