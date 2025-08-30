package ide_opening

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/hooks"
)

// TestIDEOpeningHook tests the IDE opening hook functionality.
func TestIDEOpeningHook(t *testing.T) {
	// Test successful IDE opening validation
	t.Run("successful IDE opening validation", func(t *testing.T) {
		hook := NewIDEOpeningHook()

		ctx := &hooks.HookContext{
			OperationName: "CreateWorkTree",
			Parameters: map[string]interface{}{
				"ideName": "vscode",
				"branch":  "feature/test",
			},
			Error:    nil, // Operation succeeded
			Results:  make(map[string]interface{}),
			Metadata: make(map[string]interface{}),
		}

		err := hook.PostExecute(ctx)
		if err != nil {
			t.Errorf("PostExecute should not return error: %v", err)
		}

		// Check that IDE opening information is stored in results
		if ctx.Results["ideName"] != "vscode" {
			t.Errorf("Expected IDE name 'vscode' in results, got '%v'", ctx.Results["ideName"])
		}

		if ctx.Results["worktreePath"] != "feature/test" {
			t.Errorf("Expected worktree path 'feature/test' in results, got '%v'", ctx.Results["worktreePath"])
		}

		if ctx.Results["shouldOpenIDE"] != true {
			t.Error("Expected shouldOpenIDE to be true in results")
		}
	})

	// Test no IDE opening when operation failed
	t.Run("no IDE opening when operation failed", func(t *testing.T) {
		hook := NewIDEOpeningHook()

		ctx := &hooks.HookContext{
			OperationName: "CreateWorkTree",
			Parameters: map[string]interface{}{
				"ideName": "vscode",
				"branch":  "feature/test",
			},
			Error:    &MockError{message: "operation failed"},
			Results:  make(map[string]interface{}),
			Metadata: make(map[string]interface{}),
		}

		err := hook.PostExecute(ctx)
		if err != nil {
			t.Errorf("PostExecute should not return error: %v", err)
		}

		// Check that no IDE opening information is stored when operation failed
		if ctx.Results["shouldOpenIDE"] != nil {
			t.Error("shouldOpenIDE should not be set when operation failed")
		}
	})

	// Test no IDE opening when no IDE name provided
	t.Run("no IDE opening when no IDE name provided", func(t *testing.T) {
		hook := NewIDEOpeningHook()

		ctx := &hooks.HookContext{
			OperationName: "CreateWorkTree",
			Parameters: map[string]interface{}{
				"branch": "feature/test",
			},
			Error:    nil,
			Results:  make(map[string]interface{}),
			Metadata: make(map[string]interface{}),
		}

		err := hook.PostExecute(ctx)
		if err != nil {
			t.Errorf("PostExecute should not return error: %v", err)
		}

		// Check that no IDE opening information is stored when no IDE name provided
		if ctx.Results["shouldOpenIDE"] != nil {
			t.Error("shouldOpenIDE should not be set when no IDE name provided")
		}
	})

	// Test error when branch name is empty
	t.Run("error when branch name is empty", func(t *testing.T) {
		hook := NewIDEOpeningHook()

		ctx := &hooks.HookContext{
			OperationName: "CreateWorkTree",
			Parameters: map[string]interface{}{
				"ideName": "vscode",
				"branch":  "",
			},
			Error:    nil,
			Results:  make(map[string]interface{}),
			Metadata: make(map[string]interface{}),
		}

		err := hook.PostExecute(ctx)
		if err == nil {
			t.Error("PostExecute should return error when branch name is empty")
		}
	})

	// Test RegisterForOperations method
	t.Run("RegisterForOperations", func(t *testing.T) {
		hook := NewIDEOpeningHook()

		// Create a mock CM instance
		mockCM := &MockCMForRegistration{
			registeredHooks: make(map[string][]hooks.Hook),
		}

		err := hook.RegisterForOperations(mockCM)
		if err != nil {
			t.Errorf("RegisterForOperations should not return error: %v", err)
		}

		// Check that the hook was registered for all operations
		if len(mockCM.registeredHooks[consts.CreateWorkTree]) != 1 {
			t.Error("Hook should be registered for CreateWorkTree")
		}

		if len(mockCM.registeredHooks[consts.LoadWorktree]) != 1 {
			t.Error("Hook should be registered for LoadWorktree")
		}

		if len(mockCM.registeredHooks[consts.OpenWorktree]) != 1 {
			t.Error("Hook should be registered for OpenWorktree")
		}
	})
}

// MockCMForRegistration implements the CM interface for testing RegisterForOperations.
type MockCMForRegistration struct {
	registeredHooks map[string][]hooks.Hook
}

func (m *MockCMForRegistration) RegisterHook(operation string, hook hooks.Hook) error {
	if m.registeredHooks == nil {
		m.registeredHooks = make(map[string][]hooks.Hook)
	}
	m.registeredHooks[operation] = append(m.registeredHooks[operation], hook)
	return nil
}

// MockError implements error interface for testing.
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}
