package cli

import (
	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/dependencies"
	"github.com/lerenn/code-manager/pkg/hooks"
	defaulthooks "github.com/lerenn/code-manager/pkg/hooks/default"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// newDefaultHookManager creates a default hook manager, falling back to empty manager on error.
func newDefaultHookManager() hooks.HookManagerInterface {
	hookManager, err := defaulthooks.NewDefaultHooksManager()
	if err != nil {
		// If hooks setup fails, use empty hook manager
		return hooks.NewHookManager()
	}
	return hookManager
}

// NewCodeManager creates a new CodeManager instance with the appropriate ConfigManager.
func NewCodeManager() (codemanager.CodeManager, error) {
	configManager := NewConfigManager()

	// Get config to create status manager
	config, err := configManager.GetConfigWithFallback()
	if err != nil {
		return nil, err
	}

	// Create dependencies with shared FS instance
	deps := dependencies.New().
		WithConfig(configManager).
		WithHookManager(newDefaultHookManager()).
		WithRepositoryProvider(repository.NewRepository).
		WithWorkspaceProvider(workspace.NewWorkspace).
		WithWorktreeProvider(worktree.NewWorktree)

	// Set status manager using the shared FS instance
	deps = deps.WithStatusManager(status.NewManager(deps.FS, config))

	// Validate that all dependencies are set
	if err := deps.Validate(); err != nil {
		return nil, err
	}

	return codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: deps,
	})
}
