//go:build unit

package test

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
	"github.com/lerenn/code-manager/pkg/hooks/ide/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestOpeningHook_PostExecute_Success tests successful IDE opening validation.
func TestOpeningHook_PostExecute_Success(t *testing.T) {
	hook := ide.NewOpeningHook()

	// Create a mock IDE manager
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockIDEManager := mocks.NewMockManagerInterface(ctrl)
	hook.IDEManager = mockIDEManager

	// Test successful IDE opening
	ctx := &hooks.HookContext{
		OperationName: consts.CreateWorkTree,
		Parameters: map[string]interface{}{
			"ideName":      "cursor",
			"worktreePath": "/path/to/worktree",
		},
		CM: &simpleCM{verbose: true},
	}

	// Mock IDE opening success
	mockIDEManager.EXPECT().OpenIDE("cursor", "/path/to/worktree").Return(nil)

	err := hook.PostExecute(ctx)
	assert.NoError(t, err)
}

func TestOpeningHook_PostExecute_MissingWorktreePath(t *testing.T) {
	hook := ide.NewOpeningHook()

	// Test missing worktreePath parameter
	ctx := &hooks.HookContext{
		OperationName: consts.CreateWorkTree,
		Parameters: map[string]interface{}{
			"ideName": "cursor",
			// worktreePath is missing
		},
		CM: &simpleCM{verbose: true},
	}

	err := hook.PostExecute(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ide.ErrWorktreePathRequired)
}

func TestOpeningHook_PostExecute_EmptyWorktreePath(t *testing.T) {
	hook := ide.NewOpeningHook()

	// Test empty worktreePath parameter
	ctx := &hooks.HookContext{
		OperationName: consts.CreateWorkTree,
		Parameters: map[string]interface{}{
			"ideName":      "cursor",
			"worktreePath": "",
		},
		CM: &simpleCM{verbose: true},
	}

	err := hook.PostExecute(ctx)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ide.ErrWorktreePathEmpty)
}

// simpleCM is a minimal interface implementation for testing.
type simpleCM struct {
	verbose bool
}

func (s *simpleCM) IsVerbose() bool {
	return s.verbose
}
