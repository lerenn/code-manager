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

func TestAddWorktreeToStatus_Success(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock successful worktree addition to status
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	params := StatusParams{
		RepoURL:       "github.com/test/repo",
		Branch:        "test-branch",
		WorktreePath:  "/test/repos/github.com/test/repo/worktrees/origin/test-branch",
		WorkspacePath: "",
		Remote:        "origin",
		IssueInfo:     nil,
	}

	err := repository.AddWorktreeToStatus(params)
	assert.NoError(t, err)
}

func TestAddWorktreeToStatus_RepositoryNotFound(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock worktree addition failure due to repository not found
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(status.ErrRepositoryNotFound)

	// Mock auto-add repository to status
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL("/test/repo", "origin").Return("https://github.com/test/repo.git", nil)
	mockStatus.EXPECT().AddRepository("github.com/test/repo", gomock.Any()).Return(nil)

	// Mock successful retry of worktree addition
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	params := StatusParams{
		RepoURL:       "github.com/test/repo",
		Branch:        "test-branch",
		WorktreePath:  "/test/repos/github.com/test/repo/worktrees/origin/test-branch",
		WorkspacePath: "",
		Remote:        "origin",
		IssueInfo:     nil,
	}

	err := repository.AddWorktreeToStatus(params)
	assert.NoError(t, err)
}

func TestAddWorktreeToStatus_OtherError(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock worktree addition failure with other error
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(errors.New("status file corrupted"))

	// Mock cleanup
	mockWorktree.EXPECT().CleanupDirectory("/test/repos/github.com/test/repo/worktrees/origin/test-branch").Return(nil)

	params := StatusParams{
		RepoURL:       "github.com/test/repo",
		Branch:        "test-branch",
		WorktreePath:  "/test/repos/github.com/test/repo/worktrees/origin/test-branch",
		WorkspacePath: "",
		Remote:        "origin",
		IssueInfo:     nil,
	}

	err := repository.AddWorktreeToStatus(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add worktree to status")
}

func TestAutoAddRepositoryToStatus_Success(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock Git repository check
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)

	// Mock remote URL retrieval
	mockGit.EXPECT().GetRemoteURL("/test/repo", "origin").Return("https://github.com/test/repo.git", nil)

	// Mock repository addition to status
	mockStatus.EXPECT().AddRepository("github.com/test/repo", gomock.Any()).Return(nil)

	err := repository.AutoAddRepositoryToStatus("github.com/test/repo", "/test/repo")
	assert.NoError(t, err)
}

func TestAutoAddRepositoryToStatus_NotGitRepository(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock Git repository check - not found
	mockFS.EXPECT().Exists("/test/repo/.git").Return(false, nil)

	err := repository.AutoAddRepositoryToStatus("github.com/test/repo", "/test/repo")
	assert.Error(t, err)
	assert.Equal(t, ErrNotAGitRepository, err)
}

func TestAutoAddRepositoryToStatus_NoOriginRemote(t *testing.T) {
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
		configManager:    config.NewManager("/test/config.yaml"),
		statusManager:    mockStatus,
		logger:           logger.NewNoopLogger(),
		prompt:           mockPrompt,
		worktreeProvider: func(_ worktree.NewWorktreeParams) worktree.Worktree { return mockWorktree },
		repositoryPath:   "/test/repo",
	}

	// Mock Git repository check
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)

	// Mock remote URL retrieval - no origin remote
	mockGit.EXPECT().GetRemoteURL("/test/repo", "origin").Return("", errors.New("remote not found"))

	// Mock repository addition to status (with empty remotes)
	mockStatus.EXPECT().AddRepository("github.com/test/repo", gomock.Any()).Return(nil)

	err := repository.AutoAddRepositoryToStatus("github.com/test/repo", "/test/repo")
	assert.NoError(t, err)
}
