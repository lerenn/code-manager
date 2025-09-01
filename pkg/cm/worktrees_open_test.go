//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/hooks/ide"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_OpenWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository:  mockRepository,
		HookManager: hooks.NewMockHookManagerInterface(ctrl),
		Workspace:   mockWorkspace,
		Config:      createTestConfig(),
	})

	// Override dependencies with mocks
	c := cm.(*realCM)
	c.FS = mockFS
	c.Git = mockGit
	c.repository = mockRepository

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil)

	// Mock worktree path existence check
	mockFS.EXPECT().Exists("/test/base/path/github.com/octocat/Hello-World/origin/test-branch").Return(true, nil)

	// Mock hook manager expectations
	hookManager := cm.(*realCM).hookManager.(*hooks.MockHookManagerInterface)
	hookManager.EXPECT().ExecutePreHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)
	hookManager.EXPECT().ExecutePostHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)

	// Note: IDE opening is now handled by the hook, not directly in the operation

	err := cm.OpenWorktree("test-branch", "vscode")
	assert.NoError(t, err)
}

func TestCM_OpenWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository:  mockRepository,
		HookManager: hooks.NewMockHookManagerInterface(ctrl),
		Workspace:   mockWorkspace,
		Config:      createTestConfig(),
	})

	// Override dependencies with mocks
	c := cm.(*realCM)
	c.FS = mockFS
	c.Git = mockGit
	c.repository = mockRepository

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil)

	// Mock worktree path existence check - worktree not found
	mockFS.EXPECT().Exists("/test/base/path/github.com/octocat/Hello-World/origin/test-branch").Return(false, nil)

	// Mock hook manager expectations
	hookManager := cm.(*realCM).hookManager.(*hooks.MockHookManagerInterface)
	hookManager.EXPECT().ExecutePreHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)
	hookManager.EXPECT().ExecuteErrorHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)

	err := cm.OpenWorktree("test-branch", "vscode")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
}

func TestOpenWorktree_CountsIDEOpenings(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks
	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockRepo := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockIDEManager := ide.NewMockManagerInterface(ctrl)

	// Set up FS expectations
	mockFS.EXPECT().Exists(gomock.Any()).Return(true, nil).AnyTimes()

	// Set up Git expectations
	mockGit.EXPECT().GetRepositoryName(".").Return("test-repo", nil).AnyTimes()

	// Set up repository expectations
	mockRepo.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Set up IDE manager expectations - expect exactly one call
	mockIDEManager.EXPECT().OpenIDE("vscode", gomock.Any(), gomock.Any()).Return(nil).Times(1)

	// Create a mock hook manager for testing
	mockHookManager := hooks.NewMockHookManagerInterface(ctrl)

	// Create test hook and directly override its IDE manager field
	testHook := ide.NewOpeningHook()
	testHook.IDEManager = mockIDEManager

	// Set up hook manager expectations
	mockHookManager.EXPECT().ExecutePreHooks("OpenWorktree", gomock.Any()).Return(nil).Times(1)
	mockHookManager.EXPECT().ExecutePostHooks("OpenWorktree", gomock.Any()).DoAndReturn(
		func(operation string, ctx *hooks.HookContext) error {
			// Execute the test hook manually
			return testHook.PostExecute(ctx)
		}).Times(1)

	// Create CM instance with our mock hook manager
	cmInstance := NewCMWithDependencies(NewCMParams{
		Repository:  mockRepo,
		Workspace:   mockWorkspace,
		Config:      createTestConfig(),
		HookManager: mockHookManager,
	})

	// Override dependencies with mocks
	c := cmInstance.(*realCM)
	c.FS = mockFS
	c.Git = mockGit
	c.repository = mockRepo

	// Execute OpenWorktree
	err := cmInstance.OpenWorktree("test-branch", "vscode")
	if err != nil {
		t.Errorf("OpenWorktree failed: %v", err)
	}

	// The mock will automatically verify that OpenIDE was called exactly once
	// If it was called more or fewer times, the test will fail
}
