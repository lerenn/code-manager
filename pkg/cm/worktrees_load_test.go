//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_LoadWorktree_Success(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.LoadWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.LoadWorktree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree loading
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("/test/base/path/test-repo/origin/feature-branch", nil)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_WithIDE(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repository.NewMockRepository(ctrl)
	mockWorkspace := workspace.NewMockWorkspace(ctrl)
	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockHookManager := hooks.NewMockHookManagerInterface(ctrl)

	// Create CM with mocked dependencies
	cm := NewCMWithDependencies(NewCMParams{
		Repository:  mockRepository,
		HookManager: mockHookManager,
		Workspace:   mockWorkspace,
		Config:      createTestConfig(),
	})

	// Override dependencies with mocks
	c := cm.(*realCM)
	c.fs = mockFS
	c.git = mockGit

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.LoadWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.LoadWorktree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree loading
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("/test/base/path/test-repo/origin/feature-branch", nil)

	// Note: IDE opening is now handled by the hook system, not tested here

	err := cm.LoadWorktree("origin:feature-branch", LoadWorktreeOpts{IDEName: "vscode"})
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_NewRemote(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.LoadWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.LoadWorktree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree loading with new remote
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("otheruser", "feature-branch").Return("/test/base/path/test-repo/otheruser/feature-branch", nil)

	err := cm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_SSHProtocol(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.LoadWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.LoadWorktree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree loading with SSH protocol
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("otheruser", "feature-branch").Return("/test/base/path/test-repo/otheruser/feature-branch", nil)

	err := cm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_OriginRemoteNotFound(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.LoadWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.LoadWorktree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("", ErrOriginRemoteNotFound)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOriginRemoteNotFound)
}

func TestCM_LoadWorktree_OriginRemoteInvalidURL(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.LoadWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.LoadWorktree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("", ErrOriginRemoteInvalidURL)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOriginRemoteInvalidURL)
}

func TestCM_LoadWorktree_FetchFailed(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.LoadWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.LoadWorktree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("", git.ErrFetchFailed)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, git.ErrFetchFailed)
}

func TestCM_LoadWorktree_BranchNotFound(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.LoadWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.LoadWorktree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("", git.ErrBranchNotFoundOnRemote)

	err := cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, git.ErrBranchNotFoundOnRemote)
}

func TestCM_LoadWorktree_DefaultRemote(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.LoadWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.LoadWorktree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree loading with default remote (origin)
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("", "feature-branch").Return("/test/base/path/test-repo/origin/feature-branch", nil)

	err := cm.LoadWorktree("feature-branch")
	assert.NoError(t, err)
}
