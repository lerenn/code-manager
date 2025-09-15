package cli

import (
	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/dependencies"
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

	return codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithConfig(configManager).
			WithStatusManager(status.NewManager(dependencies.New().FS, config)).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return repository.NewRepository(params)
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return workspace.NewWorkspace(params)
			}).
			WithWorktreeProvider(func(params worktree.NewWorktreeParams) worktree.Worktree {
				return worktree.NewWorktree(params)
			}),
	})
}
