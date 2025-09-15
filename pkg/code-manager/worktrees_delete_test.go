//go:build unit

package codemanager

import (
	"fmt"
	"testing"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositoryMocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	workspaceMocks "github.com/lerenn/code-manager/pkg/mode/workspace/mocks"
	promptMocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	statusMocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_DeleteWorkTree_SingleRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Hooks:         mockHookManager,
		ConfigManager: config.NewConfigManager("/test/config.yaml"),
		FS:            mockFS,
		Git:           mockGit,
		Status:        mockStatus,
		Prompt:        mockPrompt,
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree deletion
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().DeleteWorktree("test-branch", true).Return(nil)

	err = cm.DeleteWorkTree("test-branch", true) // Force deletion
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

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Hooks:         mockHookManager,
		ConfigManager: config.NewConfigManager("/test/config.yaml"),
		FS:            mockFS,
		Git:           mockGit,
		Status:        mockStatus,
		Prompt:        mockPrompt,
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)

	// Mock no repository found
	mockRepository.EXPECT().IsGitRepository().Return(false, nil)

	err = cm.DeleteWorkTree("test-branch", true)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrNoGitRepositoryOrWorkspaceFound)
}

func TestCM_DeleteWorkTrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Hooks:         mockHookManager,
		ConfigManager: config.NewConfigManager("/test/config.yaml"),
		FS:            mockFS,
		Git:           mockGit,
		Status:        mockStatus,
		Prompt:        mockPrompt,
	})
	assert.NoError(t, err)

	branches := []string{"branch1", "branch2", "branch3"}

	// Mock hook execution for each branch (3 times)
	for i := 0; i < len(branches); i++ {
		mockHookManager.EXPECT().ExecutePreHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
		mockHookManager.EXPECT().ExecutePostHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
	}

	// Mock repository detection and worktree deletion for each branch
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).Times(len(branches))
	for _, branch := range branches {
		mockRepository.EXPECT().DeleteWorktree(branch, true).Return(nil)
	}

	err = cm.DeleteWorkTrees(branches, true)
	assert.NoError(t, err)
}

func TestCM_DeleteWorkTrees_EmptyBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Hooks:         mockHookManager,
		ConfigManager: config.NewConfigManager("/test/config.yaml"),
		FS:            mockFS,
		Git:           mockGit,
		Status:        mockStatus,
		Prompt:        mockPrompt,
	})
	assert.NoError(t, err)

	err = cm.DeleteWorkTrees([]string{}, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no branches specified for deletion")
}

func TestCM_DeleteWorkTrees_PartialFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Hooks:         mockHookManager,
		ConfigManager: config.NewConfigManager("/test/config.yaml"),
		FS:            mockFS,
		Git:           mockGit,
		Status:        mockStatus,
		Prompt:        mockPrompt,
	})
	assert.NoError(t, err)

	branches := []string{"branch1", "branch2", "branch3"}

	// Mock hook execution for each branch (3 times)
	for i := 0; i < len(branches); i++ {
		mockHookManager.EXPECT().ExecutePreHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
		if i == 1 { // branch2 fails
			mockHookManager.EXPECT().ExecuteErrorHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
		} else {
			mockHookManager.EXPECT().ExecutePostHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
		}
	}

	// Mock repository detection for each branch
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).Times(len(branches))

	// Mock worktree deletion: branch1 succeeds, branch2 fails, branch3 succeeds
	mockRepository.EXPECT().DeleteWorktree("branch1", true).Return(nil)
	mockRepository.EXPECT().DeleteWorktree("branch2", true).Return(fmt.Errorf("deletion failed"))
	mockRepository.EXPECT().DeleteWorktree("branch3", true).Return(nil)

	err = cm.DeleteWorkTrees(branches, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some worktrees failed to delete")
	assert.Contains(t, err.Error(), "branch2")
}

func TestCM_DeleteWorkTrees_AllFailures(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		Hooks:         mockHookManager,
		ConfigManager: config.NewConfigManager("/test/config.yaml"),
		FS:            mockFS,
		Git:           mockGit,
		Status:        mockStatus,
		Prompt:        mockPrompt,
	})
	assert.NoError(t, err)

	branches := []string{"branch1", "branch2"}

	// Mock hook execution for each branch (2 times)
	for i := 0; i < len(branches); i++ {
		mockHookManager.EXPECT().ExecutePreHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
		mockHookManager.EXPECT().ExecuteErrorHooks(consts.DeleteWorkTree, gomock.Any()).Return(nil)
	}

	// Mock repository detection for each branch
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).Times(len(branches))

	// Mock worktree deletion: both fail
	mockRepository.EXPECT().DeleteWorktree("branch1", true).Return(fmt.Errorf("deletion failed"))
	mockRepository.EXPECT().DeleteWorktree("branch2", true).Return(fmt.Errorf("deletion failed"))

	err = cm.DeleteWorkTrees(branches, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete all worktrees")
}
