//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/hooks"
	"github.com/lerenn/code-manager/pkg/repository"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration for use in tests.
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}
}

func TestCM_Run_SingleRepository(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch").Return("/test/base/path/test-repo/origin/test-branch", nil)

	err := cm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}

func TestCM_CreateWorkTreeWithIDE(t *testing.T) {
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
	c.FS = mockFS
	c.Git = mockGit

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch").Return("/test/base/path/test-repo/origin/test-branch", nil)

	// Note: IDE opening is now handled by the hook system, not tested here

	err := cm.CreateWorkTree("test-branch", CreateWorkTreeOpts{IDEName: "vscode"})
	assert.NoError(t, err)
}

func TestCM_Run_VerboseMode(t *testing.T) {
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
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch").Return("/test/base/path/test-repo/origin/test-branch", nil)

	err := cm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}
