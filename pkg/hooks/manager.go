package hooks

import (
	"fmt"
	"sort"
	"sync"
)

//go:generate mockgen -source=manager.go -destination=mockmanager.gen.go -package=hooks

// HookManager manages hook registration and execution.
type HookManager struct {
	preHooks   map[string][]PreHook
	postHooks  map[string][]PostHook
	errorHooks map[string][]ErrorHook
	mu         sync.RWMutex
}

// HookManagerInterface defines the interface for hook management.
type HookManagerInterface interface {
	// Hook registration.
	RegisterPreHook(operation string, hook PreHook) error
	RegisterPostHook(operation string, hook PostHook) error
	RegisterErrorHook(operation string, hook ErrorHook) error

	// Hook execution.
	ExecutePreHooks(operation string, ctx *HookContext) error
	ExecutePostHooks(operation string, ctx *HookContext) error
	ExecuteErrorHooks(operation string, ctx *HookContext) error

	// Hook management.
	RemoveHook(operation, hookName string) error
	EnableHook(operation, hookName string) error
	DisableHook(operation, hookName string) error
	ListHooks(operation string) ([]Hook, error)
}

// NewHookManager creates a new HookManager instance.
func NewHookManager() HookManagerInterface {
	return &HookManager{
		preHooks:   make(map[string][]PreHook),
		postHooks:  make(map[string][]PostHook),
		errorHooks: make(map[string][]ErrorHook),
	}
}

// RegisterPreHook registers a pre-hook for a specific operation.
func (hm *HookManager) RegisterPreHook(operation string, hook PreHook) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	if hm.preHooks[operation] == nil {
		hm.preHooks[operation] = make([]PreHook, 0)
	}

	hm.preHooks[operation] = append(hm.preHooks[operation], hook)
	hm.sortHooksByPriority(operation, "pre")
	return nil
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

// RegisterErrorHook registers an error-hook for a specific operation.
func (hm *HookManager) RegisterErrorHook(operation string, hook ErrorHook) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hook == nil {
		return fmt.Errorf("hook cannot be nil")
	}

	if hm.errorHooks[operation] == nil {
		hm.errorHooks[operation] = make([]ErrorHook, 0)
	}

	hm.errorHooks[operation] = append(hm.errorHooks[operation], hook)
	hm.sortHooksByPriority(operation, "error")
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

// RemoveHook removes a hook by name from a specific operation.
func (hm *HookManager) RemoveHook(operation, hookName string) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	// Remove from pre-hooks.
	if hooks, exists := hm.preHooks[operation]; exists {
		hm.preHooks[operation] = hm.removePreHookByName(hooks, hookName)
	}

	// Remove from post-hooks.
	if hooks, exists := hm.postHooks[operation]; exists {
		hm.postHooks[operation] = hm.removePostHookByName(hooks, hookName)
	}

	// Remove from error-hooks.
	if hooks, exists := hm.errorHooks[operation]; exists {
		hm.errorHooks[operation] = hm.removeErrorHookByName(hooks, hookName)
	}

	return nil
}

// EnableHook enables a hook by name (placeholder for future implementation).
func (hm *HookManager) EnableHook(_, _ string) error {
	// TODO: Implement hook enable/disable functionality.
	return nil
}

// DisableHook disables a hook by name (placeholder for future implementation).
func (hm *HookManager) DisableHook(_, _ string) error {
	// TODO: Implement hook enable/disable functionality.
	return nil
}

// ListHooks lists all hooks for a specific operation.
func (hm *HookManager) ListHooks(operation string) ([]Hook, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var hooks []Hook

	// Add pre-hooks.
	for _, hook := range hm.preHooks[operation] {
		hooks = append(hooks, hook)
	}

	// Add post-hooks.
	for _, hook := range hm.postHooks[operation] {
		hooks = append(hooks, hook)
	}

	// Add error-hooks.
	for _, hook := range hm.errorHooks[operation] {
		hooks = append(hooks, hook)
	}

	return hooks, nil
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

func (hm *HookManager) removePreHookByName(hooks []PreHook, name string) []PreHook {
	var result []PreHook
	for _, hook := range hooks {
		if hook.Name() != name {
			result = append(result, hook)
		}
	}
	return result
}

func (hm *HookManager) removePostHookByName(hooks []PostHook, name string) []PostHook {
	var result []PostHook
	for _, hook := range hooks {
		if hook.Name() != name {
			result = append(result, hook)
		}
	}
	return result
}

func (hm *HookManager) removeErrorHookByName(hooks []ErrorHook, name string) []ErrorHook {
	var result []ErrorHook
	for _, hook := range hooks {
		if hook.Name() != name {
			result = append(result, hook)
		}
	}
	return result
}
