//go:build unit

package repository

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
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

func TestListWorktrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		fs:               mockFS,
		git:              mockGit,
		config:           config.Config{},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock GetRepositoryName
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil)

	// Mock GetRepository
	expectedWorktreeMap := map[string]status.WorktreeInfo{
		"feature-1": {Branch: "feature-1", Remote: "origin"},
		"feature-2": {Branch: "feature-2", Remote: "origin"},
	}
	expectedWorktrees := []status.WorktreeInfo{
		{Branch: "feature-1", Remote: "origin"},
		{Branch: "feature-2", Remote: "origin"},
	}
	mockRepo := &status.Repository{
		Worktrees: expectedWorktreeMap,
	}
	mockStatus.EXPECT().GetRepository("github.com/test/repo").Return(mockRepo, nil)

	result, err := repository.ListWorktrees()
	assert.NoError(t, err)
	assert.Equal(t, expectedWorktrees, result)
}

func TestListWorktrees_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		fs:               mockFS,
		git:              mockGit,
		config:           config.Config{},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock GetRepositoryName
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil)

	// Mock GetRepository to return repository not found
	mockStatus.EXPECT().GetRepository("github.com/test/repo").Return(nil, status.ErrRepositoryNotFound)

	result, err := repository.ListWorktrees()
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestListWorktrees_GetRepositoryNameError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		fs:               mockFS,
		git:              mockGit,
		config:           config.Config{},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock GetRepositoryName to return error
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("", errors.New("failed to get repository name"))

	result, err := repository.ListWorktrees()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get repository name")
}

func TestListWorktrees_StatusManagerError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		fs:               mockFS,
		git:              mockGit,
		config:           config.Config{},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock GetRepositoryName
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil)

	// Mock GetRepository to return non-ErrRepositoryNotFound error
	mockStatus.EXPECT().GetRepository("github.com/test/repo").Return(nil, errors.New("status file corrupted"))

	result, err := repository.ListWorktrees()
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "status file corrupted")
}

func TestListWorktrees_EmptyWorktrees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repository := &realRepository{
		fs:               mockFS,
		git:              mockGit,
		config:           config.Config{},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock GetRepositoryName
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil)

	// Mock GetRepository with empty worktrees
	mockRepo := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{},
	}
	mockStatus.EXPECT().GetRepository("github.com/test/repo").Return(mockRepo, nil)

	result, err := repository.ListWorktrees()
	assert.NoError(t, err)
	assert.Empty(t, result)
}
