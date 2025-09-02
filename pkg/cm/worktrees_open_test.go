//go:build unit

package cm

import (
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/repository"
	repositoryMocks "github.com/lerenn/code-manager/pkg/repository/mocks"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/workspace"
	workspaceMocks "github.com/lerenn/code-manager/pkg/workspace/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_OpenWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
		HookManager:        hooksMocks.NewMockHookManagerInterface(ctrl),
		WorkspaceProvider:  func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace },
		Config:             createTestConfig(),
		FS:                 mockFS,
		Git:                mockGit,
		Status:             mockStatus,
		Prompt:             mockPrompt,
	})
	assert.NoError(t, err)

	// Override dependencies with mocks
	c := cm.(*realCM)
	c.fs = mockFS
	c.git = mockGit
	c.repository = mockRepository

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil)

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/worktrees/github.com/octocat/Hello-World/test-branch").Return(true, nil)

	// Mock hook manager expectations
	hookManager := cm.(*realCM).hookManager.(*hooksMocks.MockHookManagerInterface)
	hookManager.EXPECT().ExecutePreHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)
	hookManager.EXPECT().ExecutePostHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)

	// Note: IDE opening is now handled by the hook, not directly in the operation

	err = cm.OpenWorktree("test-branch", "vscode")
	assert.NoError(t, err)
}

func TestCM_OpenWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
		HookManager:        hooksMocks.NewMockHookManagerInterface(ctrl),
		WorkspaceProvider:  func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace },
		Config:             createTestConfig(),
		FS:                 mockFS,
		Git:                mockGit,
		Status:             mockStatus,
		Prompt:             mockPrompt,
	})
	assert.NoError(t, err)

	// Override dependencies with mocks
	c := cm.(*realCM)
	c.fs = mockFS
	c.git = mockGit
	c.repository = mockRepository

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil)

	// Mock worktree path existence check - worktree not found
	mockFS.EXPECT().Exists("/test/base/path/worktrees/github.com/octocat/Hello-World/test-branch").Return(false, nil)

	// Mock hook manager expectations
	hookManager := cm.(*realCM).hookManager.(*hooksMocks.MockHookManagerInterface)
	hookManager.EXPECT().ExecutePreHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)
	hookManager.EXPECT().ExecuteErrorHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)

	err = cm.OpenWorktree("test-branch", "vscode")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
}

func TestOpenWorktree_CountsIDEOpenings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	// Set up FS expectations
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()

	// Set up Git expectations
	mockGit.EXPECT().GetRepositoryName(".").Return("test-repo", nil).AnyTimes()

	// Set up repository expectations
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Create a mock hook manager for testing
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	// Set up hook manager expectations
	mockHookManager.EXPECT().ExecutePreHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)
	mockHookManager.EXPECT().ExecutePostHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)

	// Create CM instance with our mock hook manager
	cmInstance, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository { return mockRepository },
		WorkspaceProvider:  func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace },
		Config:             createTestConfig(),
		FS:                 mockFS,
		Git:                mockGit,
		Status:             mockStatus,
		Prompt:             mockPrompt,
		HookManager:        mockHookManager,
	})
	assert.NoError(t, err)

	// Override dependencies with mocks
	c := cmInstance.(*realCM)
	c.fs = mockFS
	c.git = mockGit
	c.repository = mockRepository

	// Execute OpenWorktree
	err = cmInstance.OpenWorktree("test-branch", "vscode")
	assert.NoError(t, err)
}
