package ide

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/hooks"
	"go.uber.org/mock/gomock"
)

// TestOpeningHook_PostExecute_Success tests successful IDE opening validation.
func TestOpeningHook_PostExecute_Success(t *testing.T) {
	hook := NewOpeningHook()

	// Create a mock IDE manager for testing
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIDEManager := NewMockManagerInterface(ctrl)

	// Set up mock expectations
	mockIDEManager.EXPECT().OpenIDE("vscode", "feature/test", true).Return(nil)

	// Replace the real IDE manager with the mock
	hook.IDEManager = mockIDEManager

	// Create a mock CM that implements the required interfaces
	mockCM := &MockCMForPostExecute{
		verbose: true,
	}

	ctx := &hooks.HookContext{
		OperationName: "CreateWorkTree",
		Parameters: map[string]interface{}{
			"ideName": "vscode",
			"branch":  "feature/test",
		},
		Error:    nil, // Operation succeeded
		Results:  make(map[string]interface{}),
		Metadata: make(map[string]interface{}),
		CM:       mockCM,
	}

	err := hook.PostExecute(ctx)
	if err != nil {
		t.Errorf("PostExecute should not return error: %v", err)
	}
}

// TestOpeningHook_PostExecute_OperationFailed tests no IDE opening when operation failed.
func TestOpeningHook_PostExecute_OperationFailed(t *testing.T) {
	hook := NewOpeningHook()

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
}

// TestOpeningHook_PostExecute_NoIDEName tests no IDE opening when no IDE name provided.
func TestOpeningHook_PostExecute_NoIDEName(t *testing.T) {
	hook := NewOpeningHook()

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
}

// TestOpeningHook_PostExecute_EmptyBranch tests error when branch name is empty.
func TestOpeningHook_PostExecute_EmptyBranch(t *testing.T) {
	hook := NewOpeningHook()

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
}

// TestOpeningHook_RegisterForOperations tests the RegisterForOperations method.
func TestOpeningHook_RegisterForOperations(t *testing.T) {
	hook := NewOpeningHook()

	// Create a mock CM instance
	mockCM := &MockCMForRegistration{
		registeredHooks: make(map[string][]hooks.Hook),
	}

	err := hook.RegisterForOperations(mockCM.RegisterHook)
	if err != nil {
		t.Errorf("RegisterForOperations should not return error: %v", err)
	}

	// Check that the hook was registered for worktree creation operations
	if len(mockCM.registeredHooks[consts.CreateWorkTree]) != 1 {
		t.Error("Hook should be registered for CreateWorkTree")
	}

	if len(mockCM.registeredHooks[consts.LoadWorktree]) != 1 {
		t.Error("Hook should be registered for LoadWorktree")
	}

	// OpenWorktree should be registered for uniform IDE opening
	if len(mockCM.registeredHooks[consts.OpenWorktree]) != 1 {
		t.Error("Hook should be registered for OpenWorktree for uniform IDE opening")
	}
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

// MockCMForPostExecute implements the CM interface for testing PostExecute.
type MockCMForPostExecute struct {
	verbose bool
}

func (m *MockCMForPostExecute) IsVerbose() bool {
	return m.verbose
}

// MockError implements error interface for testing.
type MockError struct {
	message string
}

func (e *MockError) Error() string {
	return e.message
}
