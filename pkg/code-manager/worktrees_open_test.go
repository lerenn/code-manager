//go:build unit

package codemanager

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositoryMocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	workspaceMocks "github.com/lerenn/code-manager/pkg/mode/workspace/mocks"
	"github.com/lerenn/code-manager/pkg/prompt"
	promptMocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusMocks "github.com/lerenn/code-manager/pkg/status/mocks"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setBaselineExpectationsOpen sets common expectations for interactive flows in open tests.
func setBaselineExpectationsOpen(
	mockHookManager *hooksMocks.MockHookManagerInterface,
	mockStatus *statusMocks.MockManager,
	mockPrompt *promptMocks.MockPrompter,
	mockFS *fsmocks.MockFS,
) {
	mockHookManager.EXPECT().ExecutePreHooks(gomock.Any(), gomock.Any()).AnyTimes()
	mockHookManager.EXPECT().ExecutePostHooks(gomock.Any(), gomock.Any()).AnyTimes()
	mockHookManager.EXPECT().ExecuteErrorHooks(gomock.Any(), gomock.Any()).AnyTimes()
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{"test-repo": {}}, nil).AnyTimes()
	mockStatus.EXPECT().ListWorkspaces().Return(map[string]status.Workspace{}, nil).AnyTimes()
	mockPrompt.EXPECT().PromptSelectTarget(gomock.Any(), gomock.Any()).Return(prompt.TargetChoice{Type: prompt.TargetRepository, Name: "test-repo"}, nil).AnyTimes()
	mockFS.EXPECT().IsPathWithinBase(gomock.Any(), gomock.Any()).Return(true, nil).AnyTimes()
}

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

	// Set baseline expectations for interactive flow first
	// Note: No interactive selection since RepositoryName is provided

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.OpenWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.OpenWorktree, gomock.Any()).Return(nil)

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil).AnyTimes()

	// Mock repository validation
	mockRepository.EXPECT().ValidateRepository(gomock.Any()).Return(&repository.ValidationResult{
		RepoURL: "github.com/octocat/Hello-World",
	}, nil).AnyTimes()

	// Mock status manager to return worktree info
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil).Times(1)

	// Note: Hook expectations are handled by baseline expectations

	// Note: IDE opening is now handled by the hook, not directly in the operation

	err = cm.OpenWorktree("test-branch", "vscode", OpenWorktreeOpts{RepositoryName: "test-repo"})
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

	// Set baseline expectations for interactive flow first
	// Note: No interactive selection since RepositoryName is provided

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.OpenWorktree, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.OpenWorktree, gomock.Any()).Return(nil)

	// Mock repository detection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Mock Git to return repository URL
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/octocat/Hello-World", nil).AnyTimes()

	// Mock repository validation
	mockRepository.EXPECT().ValidateRepository(gomock.Any()).Return(&repository.ValidationResult{
		RepoURL: "github.com/octocat/Hello-World",
	}, nil).AnyTimes()

	// Mock status manager to return error (worktree not found)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(nil, status.ErrWorktreeNotFound).Times(1)

	// Note: Hook expectations are handled by baseline expectations

	err = cm.OpenWorktree("test-branch", "vscode", OpenWorktreeOpts{RepositoryName: "test-repo"})
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

	// Set baseline expectations for interactive flow first
	setBaselineExpectationsOpen(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Note: Hook expectations are handled by baseline expectations

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
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Set up Git expectations
	mockGit.EXPECT().GetRepositoryName(".").Return("test-repo", nil).AnyTimes()

	// Mock repository validation
	mockRepository.EXPECT().ValidateRepository(gomock.Any()).Return(&repository.ValidationResult{
		RepoURL: "test-repo",
	}, nil).AnyTimes()

	// Mock status manager to return worktree info
	mockStatus.EXPECT().GetWorktree("test-repo", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil).AnyTimes()
	mockStatus.EXPECT().GetWorktree("test-repo", "test-repo").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-repo",
	}, nil).AnyTimes()

	// Mock repository worktree listing
	mockRepository.EXPECT().ListWorktrees().Return([]status.WorktreeInfo{
		{Remote: "origin", Branch: "test-branch"},
	}, nil).AnyTimes()

	// Execute OpenWorktree
	err = cmInstance.OpenWorktree("test-branch", "vscode")
	assert.NoError(t, err)
}
