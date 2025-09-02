//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/cm/consts"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	promptMocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/repository"
	repositoryMocks "github.com/lerenn/code-manager/pkg/repository/mocks"
	statusMocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/workspace"
	workspaceMocks "github.com/lerenn/code-manager/pkg/workspace/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_CreateWorkTree_SingleRepository(t *testing.T) {
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
	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		HookManager: mockHookManager,
		Config:      createTestConfig(),
		FS:          mockFS,
		Git:         mockGit,
		Status:      mockStatus,
		Prompt:      mockPrompt,
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch").Return("/test/base/path/test-repo/origin/test-branch", nil)

	err = cm.CreateWorkTree("test-branch")
	assert.NoError(t, err)
}

func TestCM_CreateWorkTreeWithIDE(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepository := repositoryMocks.NewMockRepository(ctrl)
	mockWorkspace := workspaceMocks.NewMockWorkspace(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockStatus := statusMocks.NewMockManager(ctrl)
	mockPrompt := promptMocks.NewMockPrompter(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCM(NewCMParams{
		RepositoryProvider: func(params repository.NewRepositoryParams) repository.Repository {
			return mockRepository
		},
		WorkspaceProvider: func(params workspace.NewWorkspaceParams) workspace.Workspace {
			return mockWorkspace
		},
		HookManager: mockHookManager,
		Config:      createTestConfig(),
		FS:          mockFS,
		Git:         mockGit,
		Status:      mockStatus,
		Prompt:      mockPrompt,
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.CreateWorkTree, gomock.Any()).Return(nil)

	// Mock repository detection and worktree creation
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().Validate().Return(nil)
	mockRepository.EXPECT().CreateWorktree("test-branch").Return("/test/base/path/test-repo/origin/test-branch", nil)

	// Note: IDE opening is now handled by the hook system, not tested here
	err = cm.CreateWorkTree("test-branch", CreateWorkTreeOpts{IDEName: "vscode"})
	assert.NoError(t, err)
}
