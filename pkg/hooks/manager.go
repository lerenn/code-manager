package hooks

import (
	"fmt"
	"sort"
	"sync"
)

//go:generate go run go.uber.org/mock/mockgen@latest  -source=manager.go -destination=mocks/manager.gen.go -package=mocks

// HookManager manages hook registration and execution.
type HookManager struct {
	preHooks                  map[string][]PreHook
	postHooks                 map[string][]PostHook
	errorHooks                map[string][]ErrorHook
	postWorktreeCheckoutHooks map[string][]PostWorktreeCheckoutHook
	preWorktreeCreationHooks  map[string][]PreWorktreeCreationHook
	mu                        sync.RWMutex
}

// HookManagerInterface defines the interface for hook management.
type HookManagerInterface interface {
	// Hook registration.
	RegisterPostHook(operation string, hook PostHook) error
	RegisterPostWorktreeCheckoutHook(operation string, hook PostWorktreeCheckoutHook) error
	RegisterPreWorktreeCreationHook(operation string, hook PreWorktreeCreationHook) error

	// Hook execution.
	ExecutePreHooks(operation string, ctx *HookContext) error
	ExecutePostHooks(operation string, ctx *HookContext) error
	ExecuteErrorHooks(operation string, ctx *HookContext) error
	ExecutePostWorktreeCheckoutHooks(operation string, ctx *HookContext) error
	ExecutePreWorktreeCreationHooks(operation string, ctx *HookContext) error
}

// NewHookManager creates a new HookManager instance.
func NewHookManager() HookManagerInterface {
	return &HookManager{
		preHooks:                  make(map[string][]PreHook),
		postHooks:                 make(map[string][]PostHook),
		errorHooks:                make(map[string][]ErrorHook),
		postWorktreeCheckoutHooks: make(map[string][]PostWorktreeCheckoutHook),
		preWorktreeCreationHooks:  make(map[string][]PreWorktreeCreationHook),
	}
}

// RegisterPostHook registers a post-hook for a specific operation.
func (hm *HookManager) RegisterPostHook(operation string, hook PostHook) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	if hm.postHooks[operation] == nil {
		hm.postHooks[operation] = make([]PostHook, 0)
	}

	hm.postHooks[operation] = append(hm.postHooks[operation], hook)
	hm.sortHooksByPriority(operation, "post")
	return nil
}

// RegisterPostWorktreeCheckoutHook registers a post-worktree checkout hook for a specific operation.
func (hm *HookManager) RegisterPostWorktreeCheckoutHook(operation string, hook PostWorktreeCheckoutHook) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	if hm.postWorktreeCheckoutHooks[operation] == nil {
		hm.postWorktreeCheckoutHooks[operation] = make([]PostWorktreeCheckoutHook, 0)
	}

	hm.postWorktreeCheckoutHooks[operation] = append(hm.postWorktreeCheckoutHooks[operation], hook)
	hm.sortPostWorktreeCheckoutHooksByPriority(operation)
	return nil
}

// RegisterPreWorktreeCreationHook registers a pre-worktree creation hook for a specific operation.
func (hm *HookManager) RegisterPreWorktreeCreationHook(operation string, hook PreWorktreeCreationHook) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	if hm.preWorktreeCreationHooks[operation] == nil {
		hm.preWorktreeCreationHooks[operation] = make([]PreWorktreeCreationHook, 0)
	}

	hm.preWorktreeCreationHooks[operation] = append(hm.preWorktreeCreationHooks[operation], hook)
	hm.sortPreWorktreeCreationHooksByPriority(operation)
	return nil
}

// ExecutePreHooks executes all pre-hooks for a specific operation.
func (hm *HookManager) ExecutePreHooks(operation string, ctx *HookContext) error {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Execute operation-specific pre-hooks.
	for _, hook := range hm.preHooks[operation] {
		if err := hook.PreExecute(ctx); err != nil {
			return fmt.Errorf("pre-hook %s failed: %w", hook.Name(), err)
		}
	}

	return nil
}

// ExecutePostHooks executes all post-hooks for a specific operation.
func (hm *HookManager) ExecutePostHooks(operation string, ctx *HookContext) error {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Execute operation-specific post-hooks.
	for _, hook := range hm.postHooks[operation] {
		if err := hook.PostExecute(ctx); err != nil {
			return fmt.Errorf("post-hook %s failed: %w", hook.Name(), err)
		}
	}

	return nil
}

// ExecuteErrorHooks executes all error-hooks for a specific operation.
func (hm *HookManager) ExecuteErrorHooks(operation string, ctx *HookContext) error {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Execute operation-specific error-hooks.
	for _, hook := range hm.errorHooks[operation] {
		if err := hook.OnError(ctx); err != nil {
			return fmt.Errorf("error-hook %s failed: %w", hook.Name(), err)
		}
	}

	return nil
}

// ExecutePostWorktreeCheckoutHooks executes all post-worktree checkout hooks for a specific operation.
func (hm *HookManager) ExecutePostWorktreeCheckoutHooks(operation string, ctx *HookContext) error {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Execute operation-specific post-worktree checkout hooks.
	for _, hook := range hm.postWorktreeCheckoutHooks[operation] {
		if err := hook.OnPostWorktreeCheckout(ctx); err != nil {
			return fmt.Errorf("post-worktree checkout hook %s failed: %w", hook.Name(), err)
		}
	}

	return nil
}

// ExecutePreWorktreeCreationHooks executes all pre-worktree creation hooks for a specific operation.
func (hm *HookManager) ExecutePreWorktreeCreationHooks(operation string, ctx *HookContext) error {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Execute operation-specific pre-worktree creation hooks.
	for _, hook := range hm.preWorktreeCreationHooks[operation] {
		if err := hook.OnPreWorktreeCreation(ctx); err != nil {
			return fmt.Errorf("pre-worktree creation hook %s failed: %w", hook.Name(), err)
		}
	}

	return nil
}

// Helper methods for sorting and removing hooks.
func (hm *HookManager) sortHooksByPriority(operation, hookType string) {
	switch hookType {
	case "pre":
		sort.Slice(hm.preHooks[operation], func(i, j int) bool {
			return hm.preHooks[operation][i].Priority() < hm.preHooks[operation][j].Priority()
		})
	case "post":
		sort.Slice(hm.postHooks[operation], func(i, j int) bool {
			return hm.postHooks[operation][i].Priority() < hm.postHooks[operation][j].Priority()
		})
	case "error":
		sort.Slice(hm.errorHooks[operation], func(i, j int) bool {
			return hm.errorHooks[operation][i].Priority() < hm.errorHooks[operation][j].Priority()
		})
	}
}

func (hm *HookManager) sortPostWorktreeCheckoutHooksByPriority(operation string) {
	sort.Slice(hm.postWorktreeCheckoutHooks[operation], func(i, j int) bool {
		return hm.postWorktreeCheckoutHooks[operation][i].Priority() < hm.postWorktreeCheckoutHooks[operation][j].Priority()
	})
}

func (hm *HookManager) sortPreWorktreeCreationHooksByPriority(operation string) {
	sort.Slice(hm.preWorktreeCreationHooks[operation], func(i, j int) bool {
		return hm.preWorktreeCreationHooks[operation][i].Priority() < hm.preWorktreeCreationHooks[operation][j].Priority()
	})
}
