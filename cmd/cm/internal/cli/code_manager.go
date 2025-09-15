package cli

import (
	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
)

// NewCodeManager creates a new CodeManager instance with the appropriate ConfigManager.
func NewCodeManager() (codemanager.CodeManager, error) {
	configManager := NewConfigManager()
	return codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		ConfigManager: configManager,
	})
}
