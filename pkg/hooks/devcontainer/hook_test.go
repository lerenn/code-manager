//go:build unit

package devcontainer

import (
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPreWorktreeCreationHook_OnPreWorktreeCreation_WithDevcontainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	hook := &PreWorktreeCreationHook{
		fs:       mockFS,
		logger:   NewPreWorktreeCreationHook().logger,
		detector: NewDetector(mockFS),
	}

	ctx := &hooks.HookContext{
		Parameters: map[string]interface{}{
			"repoPath": "/test/repo",
		},
		Metadata: make(map[string]interface{}),
	}

	// Mock expectations
	mockFS.EXPECT().Exists("/test/repo/.devcontainer/devcontainer.json").Return(true, nil)

	err := hook.OnPreWorktreeCreation(ctx)
	assert.NoError(t, err)
	assert.True(t, ctx.Metadata["detached"].(bool))
}

func TestPreWorktreeCreationHook_OnPreWorktreeCreation_WithoutDevcontainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	hook := &PreWorktreeCreationHook{
		fs:       mockFS,
		logger:   NewPreWorktreeCreationHook().logger,
		detector: NewDetector(mockFS),
	}

	ctx := &hooks.HookContext{
		Parameters: map[string]interface{}{
			"repoPath": "/test/repo",
		},
		Metadata: make(map[string]interface{}),
	}

	// Mock expectations - both checks return false
	mockFS.EXPECT().Exists("/test/repo/.devcontainer/devcontainer.json").Return(false, nil)
	mockFS.EXPECT().Exists("/test/repo/.devcontainer.json").Return(false, nil)

	err := hook.OnPreWorktreeCreation(ctx)
	assert.NoError(t, err)
	assert.Nil(t, ctx.Metadata["detached"])
}

func TestPreWorktreeCreationHook_OnPreWorktreeCreation_NoRepoPath(t *testing.T) {
	hook := NewPreWorktreeCreationHook()

	ctx := &hooks.HookContext{
		Parameters: map[string]interface{}{},
		Metadata:   make(map[string]interface{}),
	}

	err := hook.OnPreWorktreeCreation(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository path not found")
}

func TestPreWorktreeCreationHook_Name(t *testing.T) {
	hook := NewPreWorktreeCreationHook()
	assert.Equal(t, "devcontainer-detached-worktree", hook.Name())
}

func TestPreWorktreeCreationHook_Priority(t *testing.T) {
	hook := NewPreWorktreeCreationHook()
	assert.Equal(t, 10, hook.Priority())
}
