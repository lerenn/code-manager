package cli

import (
	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/dependencies"
	defaulthooks "github.com/lerenn/code-manager/pkg/hooks/default"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
)

// NewCodeManager creates a new CodeManager instance with the appropriate ConfigManager.
func NewCodeManager() (codemanager.CodeManager, error) {
	configManager := NewConfigManager()

	// Get config to create status manager
	config, err := configManager.GetConfigWithFallback()
	if err != nil {
		return nil, err
	}

	// Create default hooks manager with IDE opening hooks
	hookManager, err := defaulthooks.NewDefaultHooksManager()
	if err != nil {
		return nil, err
	}

	return codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithConfig(configManager).
			WithStatusManager(status.NewManager(dependencies.New().FS, config)).
			WithHookManager(hookManager).
			WithRepositoryProvider(repository.NewRepository).
			WithWorkspaceProvider(workspace.NewWorkspace).
			WithWorktreeProvider(worktree.NewWorktree),
	})
}
