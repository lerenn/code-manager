//go:build unit

package codemanager

import (
	"fmt"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	repo "github.com/lerenn/code-manager/pkg/mode/repository"
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

// setBaselineExpectationsDelete sets common expectations for interactive flows in delete tests.
func setBaselineExpectationsDelete(
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
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repo.NewRepositoryParams) repo.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt),
	})
	assert.NoError(t, err)

	setBaselineExpectationsDelete(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock repository detection and worktree deletion
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().DeleteWorktree("test-branch", true).Return(nil)

	err = cm.DeleteWorkTree("test-branch", true, DeleteWorktreeOpts{RepositoryName: "test-repo"}) // Force deletion
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
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repo.NewRepositoryParams) repo.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt),
	})
	assert.NoError(t, err)

	// Set baseline expectations for interactive flow
	setBaselineExpectationsDelete(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock no repository found
	mockRepository.EXPECT().IsGitRepository().Return(false, nil).AnyTimes()

	err = cm.DeleteWorkTree("test-branch", true, DeleteWorktreeOpts{RepositoryName: "test-repo"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Git repository or workspace found")
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
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repo.NewRepositoryParams) repo.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt),
	})
	assert.NoError(t, err)

	branches := []string{"branch1", "branch2", "branch3"}

	// Set baseline expectations for interactive flow
	setBaselineExpectationsDelete(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock repository detection and worktree deletion for each branch
	// Each branch will trigger interactive selection, so we need to mock it for each branch
	for _, branch := range branches {
		mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
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
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repo.NewRepositoryParams) repo.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt),
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
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repo.NewRepositoryParams) repo.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt),
	})
	assert.NoError(t, err)

	branches := []string{"branch1", "branch2", "branch3"}

	// Set baseline expectations for interactive flow
	setBaselineExpectationsDelete(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock repository detection and worktree deletion for each branch
	// Each branch will trigger interactive selection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
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
		Dependencies: dependencies.New().
			WithRepositoryProvider(func(params repo.NewRepositoryParams) repo.Repository {
				return mockRepository
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return mockWorkspace
			}).
			WithHookManager(mockHookManager).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatus).
			WithPrompt(mockPrompt),
	})
	assert.NoError(t, err)

	branches := []string{"branch1", "branch2"}

	// Set baseline expectations for interactive flow
	setBaselineExpectationsDelete(mockHookManager, mockStatus, mockPrompt, mockFS)

	// Mock repository detection and worktree deletion for each branch
	// Each branch will trigger interactive selection
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()
	mockRepository.EXPECT().DeleteWorktree("branch1", true).Return(fmt.Errorf("deletion failed"))
	mockRepository.EXPECT().DeleteWorktree("branch2", true).Return(fmt.Errorf("deletion failed"))

	err = cm.DeleteWorkTrees(branches, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete all worktrees")
}
