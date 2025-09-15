//go:build unit

package codemanager

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositoryMocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	workspaceMocks "github.com/lerenn/code-manager/pkg/mode/workspace/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
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
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository { return mockRepository }).
			WithHookManager(mockHookManager).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace }).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt),
	})
	assert.NoError(t, err)

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).Times(1)

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil).Times(1)

	// Mock status manager to return worktree info
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil).Times(1)

	// Mock hook manager expectations
	mockHookManager.EXPECT().ExecutePreHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)
	mockHookManager.EXPECT().ExecutePostHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)

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
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository { return mockRepository }).
			WithHookManager(mockHookManager).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace }).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt),
	})
	assert.NoError(t, err)

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).Times(1)

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil).Times(1)

	// Mock status manager to return error (worktree not found)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(nil, status.ErrWorktreeNotFound).Times(1)

	// Mock hook manager expectations
	mockHookManager.EXPECT().ExecutePreHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)
	mockHookManager.EXPECT().ExecuteErrorHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)

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

	// Create a mock hook manager for testing
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	// Set up hook manager expectations
	mockHookManager.EXPECT().ExecutePreHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)
	mockHookManager.EXPECT().ExecutePostHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)

	// Create CM instance with our mock hook manager
	cmInstance, err := NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository { return mockRepository }).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace }).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt).
			WithHookManager(mockHookManager),
	})
	assert.NoError(t, err)

	// Set up repository expectations
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).Times(1)

	// Set up Git expectations
	mockGit.EXPECT().GetRepositoryName(".").Return("test-repo", nil).Times(1)

	// Mock status manager to return worktree info
	mockStatus.EXPECT().GetWorktree("test-repo", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil).Times(1)

	// Execute OpenWorktree
	err = cmInstance.OpenWorktree("test-branch", "vscode")
	assert.NoError(t, err)
}
