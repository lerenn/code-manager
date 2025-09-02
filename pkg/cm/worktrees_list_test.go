//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	"github.com/lerenn/code-manager/pkg/repository"
	repositoryMocks "github.com/lerenn/code-manager/pkg/repository/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/workspace"
	workspaceMocks "github.com/lerenn/code-manager/pkg/workspace/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_ListWorktrees_NoRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
		WorkspaceProvider:  func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace },
		Hooks:              mockHookManager,
		Config:             createTestConfig(),
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.ListWorktrees, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.ListWorktrees, gomock.Any()).Return(nil)

	// Mock repository detection to return false (no repository)
	mockRepository.EXPECT().IsGitRepository().Return(false, nil).AnyTimes()

	result, _, err := cm.ListWorktrees(false)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoGitRepositoryOrWorkspaceFound)
	assert.Nil(t, result)
}

func TestCM_ListWorktrees_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
		WorkspaceProvider:  func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace },
		Hooks:              mockHookManager,
		Config:             createTestConfig(),
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.ListWorktrees, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.ListWorktrees, gomock.Any()).Return(nil)

	// Mock repository detection and list worktrees
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	expectedWorktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "main"},
		{Remote: "origin", Branch: "feature"},
	}
	mockRepository.EXPECT().ListWorktrees().Return(expectedWorktrees, nil)

	result, projectType, err := cm.ListWorktrees(false)
	assert.NoError(t, err)
	assert.Equal(t, ProjectTypeSingleRepo, projectType)
	assert.Equal(t, expectedWorktrees, result)
}
