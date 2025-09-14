//go:build unit

package repository

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreemocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestValidateOriginRemote_Success(t *testing.T) {
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
		config:           config.Config{RepositoriesDir: "/test/repos"},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock remote existence check
	mockGit.EXPECT().RemoteExists("/test/repo", "origin").Return(true, nil)

	// Mock remote URL retrieval
	mockGit.EXPECT().GetRemoteURL("/test/repo", "origin").Return("https://github.com/test/repo.git", nil)

	// Mock host extraction (ExtractHostFromURL is a method on the repository)
	// We'll need to mock this or test it separately

	err := repository.ValidateOriginRemote()
	assert.NoError(t, err)
}

func TestValidateOriginRemote_NotFound(t *testing.T) {
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
		config:           config.Config{RepositoriesDir: "/test/repos"},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock remote existence check - remote not found
	mockGit.EXPECT().RemoteExists("/test/repo", "origin").Return(false, nil)

	err := repository.ValidateOriginRemote()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remote 'origin' not found")
}

func TestValidateOriginRemote_InvalidURL(t *testing.T) {
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
		config:           config.Config{RepositoriesDir: "/test/repos"},
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock remote existence check
	mockGit.EXPECT().RemoteExists("/test/repo", "origin").Return(true, nil)

	// Mock remote URL retrieval - invalid URL
	mockGit.EXPECT().GetRemoteURL("/test/repo", "origin").Return("invalid-url", nil)

	err := repository.ValidateOriginRemote()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has invalid URL")
}
