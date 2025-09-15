package workspace

import (
	"encoding/json"
	"fmt"
	"os"
)

// ParseFile parses a workspace configuration file.
func (w *realWorkspace) ParseFile(filename string) (Config, error) {
	w.deps.Logger.Logf("Parsing workspace file: %s", filename)

	// Read the file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read workspace file: %w", err)
	}

	// Parse the JSON content
	var workspaceConfig Config
	if err := json.Unmarshal(content, &workspaceConfig); err != nil {
		return Config{}, fmt.Errorf("failed to parse workspace file JSON: %w", err)
	}

	w.deps.Logger.Logf("Successfully parsed workspace file with %d folders", len(workspaceConfig.Folders))
	return workspaceConfig, nil
}
