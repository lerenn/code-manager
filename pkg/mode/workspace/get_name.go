package workspace

import (
	"path/filepath"
	"strings"
)

// GetName extracts the workspace name from configuration or filename.
func (w *realWorkspace) GetName(config Config, filename string) string {
	// If the configuration has a name, use it
	if config.Name != "" {
		return config.Name
	}

	// Otherwise, extract name from filename
	baseName := filepath.Base(filename)
	// Remove the .code-workspace extension
	name := strings.TrimSuffix(baseName, ".code-workspace")
	return name
}
