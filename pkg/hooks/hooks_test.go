package hooks

import (
	"testing"
)

// MockCM implements CMInterface for testing.
type MockCM struct{}

// TestHookManager tests basic hook manager functionality.
func TestHookManager(t *testing.T) {
	hm := NewHookManager()

	// Test registering a pre-hook
	preHook := &MockPreHook{name: "test-pre"}
	err := hm.RegisterPreHook("test-operation", preHook)
	if err != nil {
		t.Errorf("Failed to register pre-hook: %v", err)
	}

	// Test registering a post-hook
	postHook := &MockPostHook{name: "test-post"}
	err = hm.RegisterPostHook("test-operation", postHook)
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
		CM:            &MockCM{},
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

	// Test listing hooks
	hooks, err := hm.ListHooks("test-operation")
	if err != nil {
		t.Errorf("Failed to list hooks: %v", err)
	}
	if len(hooks) != 3 {
		t.Errorf("Expected 3 hooks, got %d", len(hooks))
	}
}

// MockPreHook implements PreHook for testing.
type MockPreHook struct {
	name string
}

func (h *MockPreHook) Name() string {
	return h.name
}

func (h *MockPreHook) Priority() int {
	return 100
}

func (h *MockPreHook) Execute(_ *HookContext) error {
	return nil
}

func (h *MockPreHook) PreExecute(_ *HookContext) error {
	return nil
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
