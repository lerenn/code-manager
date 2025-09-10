// Package defaulthooks provides default hook implementations for the code manager.
package defaulthooks

import (
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
)

// NewDefaultHooksManager creates a new default hooks manager with IDE opening hooks.
func NewDefaultHooksManager() (hooks.HookManagerInterface, error) {
	hm := hooks.NewHookManager()

	if err := ide.NewOpeningHook().RegisterForOperations(hm.RegisterPostHook); err != nil {
		return nil, err
	}

	return hm, nil
}
