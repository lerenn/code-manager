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
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreemocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestValidateWorktreeExists_Success(t *testing.T) {
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

	// Mock worktree exists in status
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil)

	err := repository.ValidateWorktreeExists("github.com/test/repo", "test-branch")
	assert.NoError(t, err)
}

func TestValidateWorktreeExists_NotFound(t *testing.T) {
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

	// Mock worktree not found in status
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(nil, status.ErrWorktreeNotFound)

	err := repository.ValidateWorktreeExists("github.com/test/repo", "test-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
	assert.Contains(t, err.Error(), "for repository github.com/test/repo branch test-branch")
}

func TestValidateWorktreeExists_StatusManagerError(t *testing.T) {
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

	// Mock status manager error
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(nil, errors.New("status error"))

	err := repository.ValidateWorktreeExists("github.com/test/repo", "test-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
	assert.Contains(t, err.Error(), "for repository github.com/test/repo branch test-branch")
}

func TestValidateWorktreeExists_NilWorktree(t *testing.T) {
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

	// Mock worktree returns nil (no error, but nil worktree)
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(nil, nil)

	err := repository.ValidateWorktreeExists("github.com/test/repo", "test-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
	assert.Contains(t, err.Error(), "for repository github.com/test/repo branch test-branch")
}

// Additional tests moved from validation_test.go

func TestValidateWorktreeExists_Success_FromValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		Dependencies: &dependencies.Dependencies{
			FS:            mockFS,
			Git:           mockGit,
			Config:        config.NewManager("/test/config.yaml"),
			StatusManager: mockStatus,
			Prompt:        mockPrompt,
		},
	})

	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(&status.WorktreeInfo{}, nil)

	err := repo.ValidateWorktreeExists("github.com/octocat/Hello-World", "test-branch")
	assert.NoError(t, err)
}

func TestValidateWorktreeExists_NotFound_FromValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	repo := NewRepository(NewRepositoryParams{
		Dependencies: &dependencies.Dependencies{
			FS:            mockFS,
			Git:           mockGit,
			Config:        config.NewManager("/test/config.yaml"),
			StatusManager: mockStatus,
			Prompt:        mockPrompt,
		},
	})

	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(nil, status.ErrWorktreeNotFound)

	err := repo.ValidateWorktreeExists("github.com/octocat/Hello-World", "test-branch")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
}
