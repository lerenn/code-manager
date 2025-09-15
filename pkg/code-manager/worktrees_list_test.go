//go:build unit

package codemanager

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/code-manager/consts"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	hooksMocks "github.com/lerenn/code-manager/pkg/hooks/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/mode/repository"
	repositoryMocks "github.com/lerenn/code-manager/pkg/mode/repository/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestListWorktrees_Success tests successful workspace worktree listing.
func TestListWorktrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()).
			WithPrompt(mockPrompt).
			WithHookManager(mockHookManager),
	}

	// Mock workspace exists with specific worktrees
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1", "feature-2"},
		Repositories: []string{"repo1", "repo2"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil)

	// Mock worktree listing - feature-1 is in repo1, feature-2 is in repo2
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
		{Remote: "origin", Branch: "feature-2"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	repo2 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-2": worktrees[1],
		},
	}
	// The logic searches for each worktree reference in all repositories
	// For feature-1: check repo1 (found), then repo2 (not found, but still called)
	// For feature-2: check repo1 (not found), then repo2 (found)
	mockStatus.EXPECT().GetRepository("repo1").Return(repo1, nil).Times(2) // Called for both feature-1 and feature-2
	mockStatus.EXPECT().GetRepository("repo2").Return(repo2, nil).Times(2) // Called for both feature-1 and feature-2

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.ListWorktrees, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.ListWorktrees, gomock.Any()).Return(nil)

	// Execute
	result, err := cm.ListWorktrees(ListWorktreesOpts{WorkspaceName: "test-workspace"})

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "feature-1", result[0].Branch)
	assert.Equal(t, "feature-2", result[1].Branch)
}

// TestListWorktrees_NotFound tests workspace worktree listing when workspace doesn't exist.
func TestListWorktrees_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()).
			WithPrompt(mockPrompt).
			WithHookManager(mockHookManager),
	}

	// Mock workspace doesn't exist
	mockStatus.EXPECT().GetWorkspace("nonexistent-workspace").Return(nil, errors.New("not found"))

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.ListWorktrees, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecuteErrorHooks(consts.ListWorktrees, gomock.Any()).Return(nil)

	// Execute
	result, err := cm.ListWorktrees(ListWorktreesOpts{WorkspaceName: "nonexistent-workspace"})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestListWorktrees_EmptyWorkspace tests workspace worktree listing when workspace has no worktrees.
func TestListWorktrees_EmptyWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()).
			WithPrompt(mockPrompt).
			WithHookManager(mockHookManager),
	}

	// Mock workspace exists but has no worktrees
	workspace := &status.Workspace{
		Worktrees:    []string{},
		Repositories: []string{},
	}
	mockStatus.EXPECT().GetWorkspace("empty-workspace").Return(workspace, nil)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.ListWorktrees, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.ListWorktrees, gomock.Any()).Return(nil)

	// Execute
	result, err := cm.ListWorktrees(ListWorktreesOpts{WorkspaceName: "empty-workspace"})

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestListWorktrees_RepositoryNotFound tests workspace worktree listing when a repository is not found.
func TestListWorktrees_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)

	cm := &realCodeManager{
		deps: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithConfig(config.NewConfigManager("/test/config.yaml")).
			WithStatusManager(mockStatus).
			WithLogger(logger.NewNoopLogger()).
			WithPrompt(mockPrompt).
			WithHookManager(mockHookManager),
	}

	// Mock workspace exists with worktrees
	workspace := &status.Workspace{
		Worktrees:    []string{"feature-1"},
		Repositories: []string{"repo1", "nonexistent-repo"},
	}
	mockStatus.EXPECT().GetWorkspace("test-workspace").Return(workspace, nil)

	// Mock repository listing - one exists, one doesn't
	worktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "feature-1"},
	}
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature-1": worktrees[0],
		},
	}
	// The logic searches for feature-1 in all repositories: repo1 (found), then nonexistent-repo (not found)
	mockStatus.EXPECT().GetRepository("repo1").Return(repo1, nil)
	mockStatus.EXPECT().GetRepository("nonexistent-repo").Return(nil, errors.New("repository not found"))

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.ListWorktrees, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.ListWorktrees, gomock.Any()).Return(nil)

	// Execute
	result, err := cm.ListWorktrees(ListWorktreesOpts{WorkspaceName: "test-workspace"})

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1) // Only the worktree from the existing repository
	assert.Equal(t, "feature-1", result[0].Branch)
}

// TestListWorktrees_RepositoryFallback tests successful repository worktree listing when no workspace is specified.
func TestListWorktrees_RepositoryFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatusManager := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockHookManager := hooksMocks.NewMockHookManagerInterface(ctrl)
	mockRepository := repositoryMocks.NewMockRepository(ctrl)

	// Create CM with mocked dependencies
	cm, err := NewCodeManager(NewCodeManagerParams{
		Dependencies: dependencies.New().
			WithFS(mockFS).
			WithGit(mockGit).
			WithStatusManager(mockStatusManager).
			WithPrompt(mockPrompt).
			WithHookManager(mockHookManager).
			WithRepositoryProvider(func(params repository.NewRepositoryParams) repository.Repository { return mockRepository }).
			WithConfig(config.NewConfigManager("/test/config.yaml")),
	})
	assert.NoError(t, err)

	// Mock hook execution
	mockHookManager.EXPECT().ExecutePreHooks(consts.ListWorktrees, gomock.Any()).Return(nil)
	mockHookManager.EXPECT().ExecutePostHooks(consts.ListWorktrees, gomock.Any()).Return(nil)

	// Mock repository detection to return single repo mode
	mockRepository.EXPECT().IsGitRepository().Return(true, nil).AnyTimes()

	// Mock repository worktree listing
	expectedWorktrees := []status.WorktreeInfo{
		{Remote: "origin", Branch: "main"},
		{Remote: "origin", Branch: "feature"},
	}
	mockRepository.EXPECT().ListWorktrees().Return(expectedWorktrees, nil)

	// Execute without workspace (should fallback to repository mode)
	result, err := cm.ListWorktrees()
	assert.NoError(t, err)
	assert.Equal(t, expectedWorktrees, result)
}
