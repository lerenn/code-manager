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

func TestDeleteWorktree_Success(t *testing.T) {
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

	// Mock repository validation
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)
	mockGit.EXPECT().Status("/test/repo").Return("On branch main", nil).AnyTimes()
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil).AnyTimes()
	mockGit.EXPECT().IsClean("/test/repo").Return(true, nil).AnyTimes()

	// Mock worktree exists validation
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil).Times(2)

	// Mock worktree path retrieval
	mockGit.EXPECT().GetWorktreePath("/test/repo", "test-branch").Return("/test/repos/github.com/test/repo/worktrees/origin/test-branch", nil)

	// Mock worktree deletion
	mockWorktree.EXPECT().Delete(worktree.DeleteParams{
		RepoURL:      "github.com/test/repo",
		Branch:       "test-branch",
		WorktreePath: "/test/repos/github.com/test/repo/worktrees/origin/test-branch",
		RepoPath:     "/test/repo",
		Force:        true,
	}).Return(nil)

	err := repository.DeleteWorktree("test-branch", true)
	assert.NoError(t, err)
}

func TestDeleteWorktree_ValidationError(t *testing.T) {
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

	// Mock repository validation failure - not a git repository
	mockFS.EXPECT().Exists("/test/repo/.git").Return(false, nil)

	err := repository.DeleteWorktree("test-branch", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "current directory is not a Git repository")
}

func TestDeleteWorktree_WorktreeNotInStatus(t *testing.T) {
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

	// Mock repository validation
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)
	mockGit.EXPECT().Status("/test/repo").Return("On branch main", nil).AnyTimes()
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil).AnyTimes()
	mockGit.EXPECT().IsClean("/test/repo").Return(true, nil).AnyTimes()

	// Mock worktree not found in status
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(nil, status.ErrWorktreeNotFound)

	err := repository.DeleteWorktree("test-branch", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found in status")
}

func TestDeleteWorktree_GetWorktreePathError(t *testing.T) {
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

	// Mock repository validation
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)
	mockGit.EXPECT().Status("/test/repo").Return("On branch main", nil).AnyTimes()
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil).AnyTimes()
	mockGit.EXPECT().IsClean("/test/repo").Return(true, nil).AnyTimes()

	// Mock worktree exists validation
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil).Times(2)

	// Mock worktree path retrieval failure
	mockGit.EXPECT().GetWorktreePath("/test/repo", "test-branch").Return("", errors.New("worktree not found"))

	// Mock RemoveFromStatus call (new behavior when worktree doesn't exist in Git)
	mockWorktree.EXPECT().RemoveFromStatus("github.com/test/repo", "test-branch").Return(nil)

	err := repository.DeleteWorktree("test-branch", false)
	assert.NoError(t, err) // Should succeed now since we remove from status
}

func TestDeleteWorktree_WorktreeDeleteError(t *testing.T) {
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

	// Mock repository validation
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)
	mockGit.EXPECT().Status("/test/repo").Return("On branch main", nil).AnyTimes()
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil).AnyTimes()
	mockGit.EXPECT().IsClean("/test/repo").Return(true, nil).AnyTimes()

	// Mock worktree exists validation
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil).Times(2)

	// Mock worktree path retrieval
	mockGit.EXPECT().GetWorktreePath("/test/repo", "test-branch").Return("/test/repos/github.com/test/repo/worktrees/origin/test-branch", nil)

	// Mock worktree deletion failure
	mockWorktree.EXPECT().Delete(worktree.DeleteParams{
		RepoURL:      "github.com/test/repo",
		Branch:       "test-branch",
		WorktreePath: "/test/repos/github.com/test/repo/worktrees/origin/test-branch",
		RepoPath:     "/test/repo",
		Force:        false,
	}).Return(errors.New("delete failed"))

	err := repository.DeleteWorktree("test-branch", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete failed")
}

func TestDeleteWorktree_ForceDeletion(t *testing.T) {
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

	// Mock repository validation
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)
	mockGit.EXPECT().Status("/test/repo").Return("On branch main", nil).AnyTimes()
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil).AnyTimes()
	mockGit.EXPECT().IsClean("/test/repo").Return(true, nil).AnyTimes()

	// Mock worktree exists validation
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil).Times(2)

	// Mock worktree path retrieval
	mockGit.EXPECT().GetWorktreePath("/test/repo", "test-branch").Return("/test/repos/github.com/test/repo/worktrees/origin/test-branch", nil)

	// Mock worktree deletion with force=true
	mockWorktree.EXPECT().Delete(worktree.DeleteParams{
		RepoURL:      "github.com/test/repo",
		Branch:       "test-branch",
		WorktreePath: "/test/repos/github.com/test/repo/worktrees/origin/test-branch",
		RepoPath:     "/test/repo",
		Force:        true,
	}).Return(nil)

	err := repository.DeleteWorktree("test-branch", true)
	assert.NoError(t, err)
}

func TestDeleteWorktree_StatusManagerError(t *testing.T) {
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

	// Mock repository validation
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil)
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil)
	mockGit.EXPECT().Status("/test/repo").Return("On branch main", nil).AnyTimes()
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil).AnyTimes()
	mockGit.EXPECT().IsClean("/test/repo").Return(true, nil).AnyTimes()

	// Mock status manager error
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "test-branch").Return(nil, errors.New("status error"))

	err := repository.DeleteWorktree("test-branch", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worktree not found in status")
}

func TestDeleteWorktree_Success_FromRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	repo := &realRepository{
		deps: &dependencies.Dependencies{
			FS:            mockFS,
			Git:           mockGit,
			Config:        config.NewManager("/test/config.yaml"),
			StatusManager: mockStatus,
			Logger:        logger.NewNoopLogger(),
			Prompt:        mockPrompt,
			WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree {
				return mockWorktree
			},
		},
		repositoryPath: ".",
	}

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree deletion
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/octocat/Hello-World", nil)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "test-branch").Return(&status.WorktreeInfo{
		Remote: "origin",
		Branch: "test-branch",
	}, nil).Times(2)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "test-branch").Return("/test/path/worktree", nil)
	mockWorktree.EXPECT().Delete(gomock.Any()).Return(nil)

	err := repo.DeleteWorktree("test-branch", true) // Force deletion
	assert.NoError(t, err)
}
