//go:build unit

package repository

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreemocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestValidateRepository_Success_NoBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		deps: &dependencies.Dependencies{
			FS:               mockFS,
			Git:              mockGit,
			Config:           config.NewManager("/test/config.yaml"),
			StatusManager:    mockStatus,
			Logger:           logger.NewNoopLogger(),
			Prompt:           mockPrompt,
			WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		},
		repositoryPath: "/test/repo",
	}

	// Mock Git repository validation
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)

	// Mock repository URL retrieval
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil)

	params := ValidationParams{
		CurrentDir: "/test/repo",
		Branch:     "", // No branch specified
	}

	result, err := repository.ValidateRepository(params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "github.com/test/repo", result.RepoURL)
	assert.Equal(t, "/test/repo", result.RepoPath)
}

func TestValidateRepository_NotGitRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		deps: &dependencies.Dependencies{
			FS:               mockFS,
			Git:              mockGit,
			Config:           config.NewManager("/test/config.yaml"),
			StatusManager:    mockStatus,
			Logger:           logger.NewNoopLogger(),
			Prompt:           mockPrompt,
			WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		},
		repositoryPath: "/test/repo",
	}

	// Mock Git repository validation - not a Git repository
	mockFS.EXPECT().Exists("/test/repo/.git").Return(false, nil)

	params := ValidationParams{
		CurrentDir: "/test/repo",
		Branch:     "",
	}

	result, err := repository.ValidateRepository(params)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "current directory is not a Git repository")
}

func TestValidateRepository_WorktreeAlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		deps: &dependencies.Dependencies{
			FS:               mockFS,
			Git:              mockGit,
			Config:           config.NewManager("/test/config.yaml"),
			StatusManager:    mockStatus,
			Logger:           logger.NewNoopLogger(),
			Prompt:           mockPrompt,
			WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		},
		repositoryPath: "/test/repo",
	}

	// Mock Git repository validation
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)

	// Mock repository URL retrieval
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil)

	// Mock worktree existence check - worktree already exists
	existingWorktree := &status.WorktreeInfo{Branch: "test-branch", Remote: "origin"}
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(existingWorktree, nil)

	params := ValidationParams{
		CurrentDir: "/test/repo",
		Branch:     "test-branch",
	}

	result, err := repository.ValidateRepository(params)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "worktree already exists")
}
