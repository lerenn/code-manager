//go:build unit

package codemanager

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	configmocks "github.com/lerenn/code-manager/pkg/config/mocks"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositorymocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	workspacemocks "github.com/lerenn/code-manager/pkg/mode/workspace/mocks"
	"github.com/lerenn/code-manager/pkg/prompt"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreemocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/mock/gomock"
)

func TestPromptSelectTarget_NoChoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockConfig := configmocks.NewMockManager(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)

	// Setup mocks
	testConfig := config.Config{
		RepositoriesDir: "/test/repos",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil)
	mockStatus.EXPECT().ListRepositories().Return(map[string]status.Repository{}, nil)
	mockStatus.EXPECT().ListWorkspaces().Return(map[string]status.Workspace{}, nil)

	// Create CM instance
	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(mockConfig).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()).
			WithPrompt(mockPrompt).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return repositorymocks.NewMockRepository(ctrl)
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return workspacemocks.NewMockWorkspace(ctrl)
			}).
			WithWorktreeProvider(func(params worktree.NewWorktreeParams) worktree.Worktree {
				return worktreemocks.NewMockWorktree(ctrl)
			}),
	}

	// Test with no choices available
	_, err := cm.promptSelectTargetOnly()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no repositories or workspaces available")
}

func TestPromptSelectTarget_WithChoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockConfig := configmocks.NewMockManager(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)

	// Setup repositories and workspaces
	repositories := map[string]status.Repository{
		"test-repo": {
			Path: "/path/to/test-repo",
		},
	}
	workspaces := map[string]status.Workspace{
		"test-workspace": {
			Repositories: []string{"test-repo"},
		},
	}

	// Setup mocks
	testConfig := config.Config{
		RepositoriesDir: "/test/repos",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil)
	mockFS.EXPECT().IsPathWithinBase("/test/repos", "/path/to/test-repo").Return(true, nil)
	mockStatus.EXPECT().ListRepositories().Return(repositories, nil)
	mockStatus.EXPECT().ListWorkspaces().Return(workspaces, nil)

	// Mock the prompt selection
	expectedChoice := prompt.TargetChoice{
		Type: prompt.TargetRepository,
		Name: "test-repo",
	}
	mockPrompt.EXPECT().PromptSelectTarget(mock.MatchedBy(func(choices []prompt.TargetChoice) bool {
		return len(choices) == 2 && choices[0].Name == "test-repo" && choices[1].Name == "test-workspace"
	}), false).Return(expectedChoice, nil)

	// Create CM instance
	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(mockConfig).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()).
			WithPrompt(mockPrompt).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return repositorymocks.NewMockRepository(ctrl)
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return workspacemocks.NewMockWorkspace(ctrl)
			}).
			WithWorktreeProvider(func(params worktree.NewWorktreeParams) worktree.Worktree {
				return worktreemocks.NewMockWorktree(ctrl)
			}),
	}

	// Test selection
	result, err := cm.promptSelectTargetOnly()
	assert.NoError(t, err)
	assert.Equal(t, "test-repo", result.Name)
	assert.Equal(t, "repository", result.Type)
}

func TestBuildTargetChoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockConfig := configmocks.NewMockManager(ctrl)
	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)

	// Setup repositories and workspaces
	repositories := map[string]status.Repository{
		"alpha-repo": {
			Path: "/path/to/alpha-repo",
		},
		"beta-repo": {
			Path: "/path/to/beta-repo",
		},
	}
	workspaces := map[string]status.Workspace{
		"gamma-workspace": {
			Repositories: []string{"alpha-repo"},
		},
		"delta-workspace": {
			Repositories: []string{"beta-repo"},
		},
	}

	// Setup mocks
	testConfig := config.Config{
		RepositoriesDir: "/test/repos",
		WorkspacesDir:   "/test/workspaces",
		StatusFile:      "/test/status.yaml",
	}
	mockConfig.EXPECT().GetConfigWithFallback().Return(testConfig, nil)
	mockFS.EXPECT().IsPathWithinBase("/test/repos", "/path/to/alpha-repo").Return(true, nil)
	mockFS.EXPECT().IsPathWithinBase("/test/repos", "/path/to/beta-repo").Return(true, nil)
	mockStatus.EXPECT().ListRepositories().Return(repositories, nil)
	mockStatus.EXPECT().ListWorkspaces().Return(workspaces, nil)

	// Create CM instance
	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(mockConfig).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()).
			WithPrompt(promptmocks.NewMockPrompter(ctrl)).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository {
				return repositorymocks.NewMockRepository(ctrl)
			}).
			WithWorkspaceProvider(func(params workspace.NewWorkspaceParams) workspace.Workspace {
				return workspacemocks.NewMockWorkspace(ctrl)
			}).
			WithWorktreeProvider(func(params worktree.NewWorktreeParams) worktree.Worktree {
				return worktreemocks.NewMockWorktree(ctrl)
			}),
	}

	// Test building choices
	choices, err := cm.buildTargetChoices(false, "")
	assert.NoError(t, err)
	assert.Len(t, choices, 4) // 2 repos + 2 workspaces

	// Check that repositories come first (sorted by type)
	assert.Equal(t, prompt.TargetRepository, choices[0].Type)
	assert.Equal(t, "alpha-repo", choices[0].Name)
	assert.Equal(t, prompt.TargetRepository, choices[1].Type)
	assert.Equal(t, "beta-repo", choices[1].Name)

	// Check that workspaces come after repositories
	assert.Equal(t, prompt.TargetWorkspace, choices[2].Type)
	assert.Equal(t, "delta-workspace", choices[2].Name)
	assert.Equal(t, prompt.TargetWorkspace, choices[3].Type)
	assert.Equal(t, "gamma-workspace", choices[3].Name)
}
