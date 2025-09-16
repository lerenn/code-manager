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

func TestValidateGitConfiguration_Success(t *testing.T) {
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

	// Mock successful git configuration check
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("main", nil)

	err := repository.ValidateGitConfiguration("/test/repo")
	assert.NoError(t, err)
}

func TestValidateGitConfiguration_WithFeatureBranch(t *testing.T) {
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

	// Mock successful git configuration check with feature branch
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("feature/new-feature", nil)

	err := repository.ValidateGitConfiguration("/test/repo")
	assert.NoError(t, err)
}

func TestValidateGitConfiguration_DetachedHead(t *testing.T) {
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

	// Mock git configuration check in detached HEAD state
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("HEAD", nil)

	err := repository.ValidateGitConfiguration("/test/repo")
	assert.NoError(t, err)
}

func TestValidateGitConfiguration_GitNotAvailable(t *testing.T) {
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

	// Mock git command failure - git not available
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("", errors.New("git: command not found"))

	err := repository.ValidateGitConfiguration("/test/repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git configuration validation failed")
	assert.Contains(t, err.Error(), "git: command not found")
}

func TestValidateGitConfiguration_NotGitRepository(t *testing.T) {
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

	// Mock git command failure - not a git repository
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("", errors.New("not a git repository"))

	err := repository.ValidateGitConfiguration("/test/repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git configuration validation failed")
	assert.Contains(t, err.Error(), "not a git repository")
}

func TestValidateGitConfiguration_PermissionDenied(t *testing.T) {
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

	// Mock git command failure - permission denied
	mockGit.EXPECT().GetCurrentBranch("/test/repo").Return("", errors.New("permission denied"))

	err := repository.ValidateGitConfiguration("/test/repo")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git configuration validation failed")
	assert.Contains(t, err.Error(), "permission denied")
}

func TestValidateGitConfiguration_EmptyWorkDir(t *testing.T) {
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

	// Mock git command failure - empty work directory
	mockGit.EXPECT().GetCurrentBranch("").Return("", errors.New("fatal: not a git repository"))

	err := repository.ValidateGitConfiguration("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git configuration validation failed")
	assert.Contains(t, err.Error(), "fatal: not a git repository")
}

func TestValidateGitConfiguration_DifferentWorkDir(t *testing.T) {
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

	// Mock successful git configuration check with different work directory
	mockGit.EXPECT().GetCurrentBranch("/different/path").Return("develop", nil)

	err := repository.ValidateGitConfiguration("/different/path")
	assert.NoError(t, err)
}
