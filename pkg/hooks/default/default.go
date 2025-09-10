// Package defaulthooks provides default hook implementations for the code manager.
package defaulthooks

import (
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/hooks/gitcrypt"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
)

// NewDefaultHooksManager creates a new default hooks manager with IDE opening hooks and git-crypt support.
func NewDefaultHooksManager() (hooks.HookManagerInterface, error) {
	hm := hooks.NewHookManager()

	// Register IDE opening hook
	if err := ide.NewOpeningHook().RegisterForOperations(hm.RegisterPostHook); err != nil {
		return nil, err
	}

	// Register git-crypt worktree checkout hook
	gitCryptHook := gitcrypt.NewWorktreeCheckoutHook()
	if err := gitCryptHook.RegisterForOperations(hm.RegisterWorktreeCheckoutHook); err != nil {
		return nil, err
	}

	return hm, nil
}
