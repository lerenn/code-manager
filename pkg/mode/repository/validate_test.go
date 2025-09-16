//go:build unit

package repository

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
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

// createTestConfig creates a test configuration.
func createTestConfig() config.Config {
	return config.Config{
		RepositoriesDir: "/test/repositories/path",
		WorkspacesDir:   "/test/workspaces/path",
		StatusFile:      "/test/status.yaml",
	}
}

func TestValidate_Success(t *testing.T) {
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

	// Mock IsGitRepository to return true
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)

	// Mock ValidateGitStatus
	mockGit.EXPECT().Status("/test/repo").Return("On branch main\nnothing to commit, working tree clean", nil)

	// Mock ValidateGitConfiguration
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("main", nil)

	err := repository.Validate()
	assert.NoError(t, err)
}

func TestValidate_NotGitRepository(t *testing.T) {
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

	// Mock IsGitRepository to return false
	mockFS.EXPECT().Exists("/test/repo/.git").Return(false, nil)

	err := repository.Validate()
	assert.Error(t, err)
	assert.Equal(t, ErrGitRepositoryNotFound, err)
}

func TestValidate_GitStatusError(t *testing.T) {
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

	// Mock IsGitRepository to return true
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)

	// Mock ValidateGitStatus to return error
	mockGit.EXPECT().Status("/test/repo").Return("", errors.New("git status failed"))

	err := repository.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git status failed")
}

func TestValidate_GitConfigurationError(t *testing.T) {
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

	// Mock IsGitRepository to return true
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)

	// Mock ValidateGitStatus
	mockGit.EXPECT().Status("/test/repo").Return("On branch main\nnothing to commit, working tree clean", nil)

	// Mock ValidateGitConfiguration to return error
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("", errors.New("git config failed"))

	err := repository.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git config failed")
}

func TestValidate_IsGitRepositoryError(t *testing.T) {
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

	// Mock IsGitRepository to return error
	mockFS.EXPECT().Exists("/test/repo/.git").Return(false, errors.New("filesystem error"))

	err := repository.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "filesystem error")
}

func TestValidate_WorktreeRepository(t *testing.T) {
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

	// Mock worktree repository (.git is a file, not directory)
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(false, nil)
	mockFS.EXPECT().ReadFile("/test/repo/.git").Return([]byte("gitdir: /path/to/main/repo/.git/worktrees/worktree-name"), nil)

	// Mock ValidateGitStatus
	mockGit.EXPECT().Status("/test/repo").Return("On branch main\nnothing to commit, working tree clean", nil)

	// Mock ValidateGitConfiguration
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("main", nil)

	err := repository.Validate()
	assert.NoError(t, err)
}

func TestValidate_WithUncommittedChanges(t *testing.T) {
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

	// Mock IsGitRepository to return true
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)

	// Mock ValidateGitStatus with uncommitted changes
	mockGit.EXPECT().Status("/test/repo").Return("On branch main\nChanges not staged for commit:\n\tmodified:   file.txt", nil)

	// Mock ValidateGitConfiguration
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("main", nil)

	err := repository.Validate()
	assert.NoError(t, err)
}

func TestValidate_DetachedHead(t *testing.T) {
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

	// Mock IsGitRepository to return true
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)

	// Mock ValidateGitStatus
	mockGit.EXPECT().Status("/test/repo").Return("HEAD detached at abc1234\nnothing to commit, working tree clean", nil)

	// Mock ValidateGitConfiguration with detached head
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("", nil)

	err := repository.Validate()
	assert.NoError(t, err)
}
