//go:build unit

package repository

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/dependencies"
	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/lerenn/code-manager/pkg/git"
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

func TestLoadWorktree_Success(t *testing.T) {
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
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil).AnyTimes()

	// Mock remote validation and management (origin remote exists)
	mockGit.EXPECT().RemoteExists("/test/repo", "origin").Return(true, nil).AnyTimes()
	mockGit.EXPECT().GetRemoteURL("/test/repo", "origin").Return("https://github.com/test/repo.git", nil).AnyTimes()

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote("/test/repo", "origin").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   "/test/repo",
		RemoteName: "origin",
		Branch:     "feature-branch",
	}).Return(true, nil)

	// Mock worktree creation (called by CreateWorktree)
	mockGit.EXPECT().GetRepositoryName("/test/repo").Return("github.com/test/repo", nil)
	mockStatus.EXPECT().GetWorktree("github.com/test/repo", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean("/test/repo").Return(true, nil)
	mockWorktree.EXPECT().BuildPath("github.com/test/repo", "origin", "feature-branch").Return("/test/repos/github.com/test/repo/worktrees/origin/feature-branch")
	mockWorktree.EXPECT().ValidateCreation(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().Create(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().CheckoutBranch("/test/repos/github.com/test/repo/worktrees/origin/feature-branch", "feature-branch").Return(nil)
	mockGit.EXPECT().SetUpstreamBranch("/test/repos/github.com/test/repo/worktrees/origin/feature-branch", "origin", "feature-branch").Return(nil)
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	worktreePath, err := repository.LoadWorktree("origin", "feature-branch")
	assert.NoError(t, err)
	assert.Equal(t, "/test/repos/github.com/test/repo/worktrees/origin/feature-branch", worktreePath)
}

func TestLoadWorktree_NotGitRepository(t *testing.T) {
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

	// Mock Git repository validation - not a git repository
	mockFS.EXPECT().Exists("/test/repo/.git").Return(false, nil)

	worktreePath, err := repository.LoadWorktree("origin", "feature-branch")
	assert.Error(t, err)
	assert.Empty(t, worktreePath)
	assert.Equal(t, ErrGitRepositoryNotFound, err)
}

func TestLoadWorktree_RemoteNotFound(t *testing.T) {
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

	// Mock remote validation - remote not found
	mockGit.EXPECT().RemoteExists("/test/repo", "upstream").Return(false, nil)

	worktreePath, err := repository.LoadWorktree("upstream", "feature-branch")
	assert.Error(t, err)
	assert.Empty(t, worktreePath)
	assert.Contains(t, err.Error(), "remote 'upstream' not found")
}

func TestLoadWorktree_BranchNotFoundOnRemote(t *testing.T) {
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
	mockFS.EXPECT().Exists("/test/repo/.git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir("/test/repo/.git").Return(true, nil).AnyTimes()

	// Mock remote validation and management
	mockGit.EXPECT().RemoteExists("/test/repo", "origin").Return(true, nil).AnyTimes()
	mockGit.EXPECT().GetRemoteURL("/test/repo", "origin").Return("https://github.com/test/repo.git", nil).AnyTimes()

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote("/test/repo", "origin").Return(nil)

	// Mock branch existence check - branch not found
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   "/test/repo",
		RemoteName: "origin",
		Branch:     "nonexistent-branch",
	}).Return(false, nil)

	worktreePath, err := repository.LoadWorktree("origin", "nonexistent-branch")
	assert.Error(t, err)
	assert.Empty(t, worktreePath)
	assert.Contains(t, err.Error(), "branch 'nonexistent-branch' not found on remote 'origin'")
}

// Additional test moved from repository_test.go

func TestLoadWorktree_Success_FromRepository(t *testing.T) {
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

	// Mock Git repository validation (called by LoadWorktree and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/octocat/Hello-World.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "origin").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(gomock.Any()).Return(true, nil)

	// Mock worktree creation (reusing existing logic) - called by CreateWorktree
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/octocat/Hello-World", nil)
	mockStatus.EXPECT().GetWorktree("github.com/octocat/Hello-World", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockWorktree.EXPECT().BuildPath("github.com/octocat/Hello-World", "origin", "feature-branch").Return("/test/path/github.com/octocat/Hello-World/origin/feature-branch")
	mockWorktree.EXPECT().ValidateCreation(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().Create(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().CheckoutBranch("/test/path/github.com/octocat/Hello-World/origin/feature-branch", "feature-branch").Return(nil)
	mockGit.EXPECT().SetUpstreamBranch("/test/path/github.com/octocat/Hello-World/origin/feature-branch", "origin", "feature-branch").Return(nil)
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	worktreePath, err := repo.LoadWorktree("origin", "feature-branch")
	assert.NoError(t, err)
	assert.NotEmpty(t, worktreePath)
}
