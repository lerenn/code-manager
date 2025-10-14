//go:build unit

package codemanager

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/lerenn/code-manager/pkg/git"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositoryMocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	workspaceMocks "github.com/lerenn/code-manager/pkg/mode/workspace/mocks"
	"github.com/lerenn/code-manager/pkg/prompt"
	promptMocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusMocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setBaselineExpectationsLoad sets common expectations for interactive flows in load tests.
func setBaselineExpectationsLoad(
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

func TestCM_LoadWorktree_Success(t *testing.T) {
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
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository { return mockRepository }).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace { return mockWorkspace }).
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt),
	})
	assert.NoError(t, err)

	setBaselineExpectationsLoad(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock interactive selection to return a repository
	// Note: baseline expectations handle the PromptSelectTarget and hook calls

	// Mock repository detection and worktree loading
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("/test/base/path/test-repo/origin/feature-branch", nil)

	err = cm.LoadWorktree("origin:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_WithIDE(t *testing.T) {
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

	// Set baseline expectations for interactive flow
	setBaselineExpectationsLoad(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock hook execution - interactive selection calls ListRepositories first, then PromptSelectTarget
	// Note: baseline expectations handle the hook calls

	// Mock repository detection and worktree loading
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("/test/base/path/test-repo/origin/feature-branch", nil)

	// Note: IDE opening is now handled by the hook system, not tested here
	err = cm.LoadWorktree("origin:feature-branch", LoadWorktreeOpts{IDEName: "vscode"})
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_NewRemote(t *testing.T) {
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

	// Set baseline expectations for interactive flow
	setBaselineExpectationsLoad(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock hook execution - interactive selection calls ListRepositories first, then PromptSelectTarget
	// Note: baseline expectations handle the hook calls

	// Mock repository detection and worktree loading with new remote
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("otheruser", "feature-branch").Return("/test/base/path/test-repo/otheruser/feature-branch", nil)

	err = cm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_SSHProtocol(t *testing.T) {
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

	// Set baseline expectations for interactive flow
	setBaselineExpectationsLoad(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock hook execution - interactive selection calls ListRepositories first, then PromptSelectTarget
	// Note: baseline expectations handle the hook calls
	// Note: baseline expectations handle the hook calls

	// Mock repository detection and worktree loading with SSH protocol
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("otheruser", "feature-branch").Return("/test/base/path/test-repo/otheruser/feature-branch", nil)

	err = cm.LoadWorktree("otheruser:feature-branch")
	assert.NoError(t, err)
}

func TestCM_LoadWorktree_OriginRemoteNotFound(t *testing.T) {
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

	// Set baseline expectations for interactive flow
	setBaselineExpectationsLoad(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock hook execution
	// Note: baseline expectations handle the hook calls

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("", ErrOriginRemoteNotFound)

	err = cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOriginRemoteNotFound)
}

func TestCM_LoadWorktree_OriginRemoteInvalidURL(t *testing.T) {
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

	// Set baseline expectations for interactive flow
	setBaselineExpectationsLoad(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock hook execution
	// Note: baseline expectations handle the hook calls

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("", ErrOriginRemoteInvalidURL)

	err = cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrOriginRemoteInvalidURL)
}

func TestCM_LoadWorktree_FetchFailed(t *testing.T) {
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

	// Set baseline expectations for interactive flow
	setBaselineExpectationsLoad(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock hook execution
	// Note: baseline expectations handle the hook calls

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("", git.ErrFetchFailed)

	err = cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, git.ErrFetchFailed)
}

func TestCM_LoadWorktree_BranchNotFound(t *testing.T) {
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

	// Set baseline expectations for interactive flow
	setBaselineExpectationsLoad(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock hook execution
	// Note: baseline expectations handle the hook calls

	// Mock repository detection and worktree loading to return an error
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("origin", "feature-branch").Return("", git.ErrBranchNotFoundOnRemote)

	err = cm.LoadWorktree("origin:feature-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, git.ErrBranchNotFoundOnRemote)
}

func TestCM_LoadWorktree_DefaultRemote(t *testing.T) {
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

	// Set baseline expectations for interactive flow
	setBaselineExpectationsLoad(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock hook execution - interactive selection calls ListRepositories first, then PromptSelectTarget
	// Note: baseline expectations handle the hook calls
	// Note: baseline expectations handle the hook calls

	// Mock repository detection and worktree loading with default remote (origin)
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().LoadWorktree("", "feature-branch").Return("/test/base/path/test-repo/origin/feature-branch", nil)

	err = cm.LoadWorktree("feature-branch")
	assert.NoError(t, err)
}
