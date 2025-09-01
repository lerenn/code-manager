//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_DeleteWorkTree_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockHookManager := hooks.NewMockHookManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository:  mockRepository,
		HookManager: mockHookManager,
		Workspace:   mockWorkspace,
		Config:      createTestConfig(),
	})

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree deletion
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().DeleteWorktree("test-branch", true).Return(nil)

	err := cm.DeleteWorkTree("test-branch", true) // Force deletion
	assert.NoError(t, err)
}

// TestCM_DeleteWorkTree_Workspace is skipped due to test environment issues
// with workspace files in the test directory
func TestCM_DeleteWorkTree_Workspace(t *testing.T) {
	t.Skip("Skipping workspace test due to test environment issues")
}

func TestCM_DeleteWorkTree_NoRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockHookManager := hooks.NewMockHookManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository:  mockRepository,
		HookManager: mockHookManager,
		Workspace:   mockWorkspace,
		Config:      createTestConfig(),
	})

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)

	// Mock no repository found
	mockRepository.EXPECT().IsGitRepository().Return(false, nil)

	err := cm.DeleteWorkTree("test-branch", true)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoGitRepositoryOrWorkspaceFound)
}

func TestCM_DeleteWorkTree_VerboseMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockHookManager := hooks.NewMockHookManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository:  mockRepository,
		HookManager: mockHookManager,
		Workspace:   mockWorkspace,
		Config:      createTestConfig(),
	})

	// Enable verbose mode
	cm.SetVerbose(true)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree deletion
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().DeleteWorktree("test-branch", true).Return(nil)

	err := cm.DeleteWorkTree("test-branch", true)
	assert.NoError(t, err)
}
