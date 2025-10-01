package hooks

import (
	"fmt"
	"sort"
	"sync"
)

//go:generate go run go.uber.org/mock/mockgen@latest  -source=manager.go -destination=mocks/manager.gen.go -package=mocks

// HookManager manages hook registration and execution.
type HookManager struct {
	preHooks              map[string][]PreHook
	postHooks             map[string][]PostHook
	errorHooks            map[string][]ErrorHook
	worktreeCheckoutHooks map[string][]WorktreeCheckoutHook
	mu                    sync.RWMutex
}

// HookManagerInterface defines the interface for hook management.
type HookManagerInterface interface {
	// Hook registration.
	RegisterPostHook(operation string, hook PostHook) error
	RegisterWorktreeCheckoutHook(operation string, hook WorktreeCheckoutHook) error

	// Hook execution.
	ExecutePreHooks(operation string, ctx *HookContext) error
	ExecutePostHooks(operation string, ctx *HookContext) error
	ExecuteErrorHooks(operation string, ctx *HookContext) error
	ExecuteWorktreeCheckoutHooks(operation string, ctx *HookContext) error
}

// NewHookManager creates a new HookManager instance.
func NewHookManager() HookManagerInterface {
	return &HookManager{
		preHooks:              make(map[string][]PreHook),
		postHooks:             make(map[string][]PostHook),
		errorHooks:            make(map[string][]ErrorHook),
		worktreeCheckoutHooks: make(map[string][]WorktreeCheckoutHook),
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

// RegisterWorktreeCheckoutHook registers a worktree checkout hook for a specific operation.
func (hm *HookManager) RegisterWorktreeCheckoutHook(operation string, hook WorktreeCheckoutHook) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	if hm.worktreeCheckoutHooks[operation] == nil {
		hm.worktreeCheckoutHooks[operation] = make([]WorktreeCheckoutHook, 0)
	}

	hm.worktreeCheckoutHooks[operation] = append(hm.worktreeCheckoutHooks[operation], hook)
	hm.sortWorktreeCheckoutHooksByPriority(operation)
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

// ExecuteWorktreeCheckoutHooks executes all worktree checkout hooks for a specific operation.
func (hm *HookManager) ExecuteWorktreeCheckoutHooks(operation string, ctx *HookContext) error {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	// Execute operation-specific worktree checkout hooks.
	for _, hook := range hm.worktreeCheckoutHooks[operation] {
		if err := hook.OnWorktreeCheckout(ctx); err != nil {
			return fmt.Errorf("worktree checkout hook %s failed: %w", hook.Name(), err)
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

func (hm *HookManager) sortWorktreeCheckoutHooksByPriority(operation string) {
	sort.Slice(hm.worktreeCheckoutHooks[operation], func(i, j int) bool {
		return hm.worktreeCheckoutHooks[operation][i].Priority() < hm.worktreeCheckoutHooks[operation][j].Priority()
	})
}
