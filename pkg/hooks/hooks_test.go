package hooks

import (
	"testing"
)

// TestHookManager tests basic hook manager functionality.
func TestHookManager(t *testing.T) {
	hm := NewHookManager()

	// Test registering a post-hook
	postHook := &MockPostHook{name: "test-post"}
	err := hm.RegisterPostHook("test-operation", postHook)
	if err != nil {
		t.Errorf("Failed to register post-hook: %v", err)
	}

	// Test registering a worktree checkout hook
	worktreeCheckoutHook := &MockWorktreeCheckoutHook{name: "test-worktree-checkout"}
	err = hm.RegisterWorktreeCheckoutHook("test-operation", worktreeCheckoutHook)
	if err != nil {
		t.Errorf("Failed to register worktree checkout hook: %v", err)
	}

	// Test hook execution
	ctx := &HookContext{
		OperationName: "test-operation",
		Parameters:    map[string]interface{}{"test": "value"},
		Results:       map[string]interface{}{"success": true},
		Metadata:      make(map[string]interface{}),
	}

	// Execute pre-hooks
	err = hm.ExecutePreHooks("test-operation", ctx)
	if err != nil {
		t.Errorf("Failed to execute pre-hooks: %v", err)
	}

	// Execute post-hooks
	err = hm.ExecutePostHooks("test-operation", ctx)
	if err != nil {
		t.Errorf("Failed to execute post-hooks: %v", err)
	}

	// Execute worktree checkout hooks
	err = hm.ExecuteWorktreeCheckoutHooks("test-operation", ctx)
	if err != nil {
		t.Errorf("Failed to execute worktree checkout hooks: %v", err)
	}
}

// MockPostHook implements PostHook for testing.
type MockPostHook struct {
	name string
}

func (h *MockPostHook) Name() string {
	return h.name
}

func (h *MockPostHook) Priority() int {
	return 200
}

func (h *MockPostHook) Execute(_ *HookContext) error {
	return nil
}

func (h *MockPostHook) PostExecute(_ *HookContext) error {
	return nil
}

// MockWorktreeCheckoutHook implements WorktreeCheckoutHook for testing.
type MockWorktreeCheckoutHook struct {
	name string
}

func (h *MockWorktreeCheckoutHook) Name() string {
	return h.name
}

func (h *MockWorktreeCheckoutHook) Priority() int {
	return 150
}

func (h *MockWorktreeCheckoutHook) Execute(_ *HookContext) error {
	return nil
}

func (h *MockWorktreeCheckoutHook) OnWorktreeCheckout(_ *HookContext) error {
	return nil
}
