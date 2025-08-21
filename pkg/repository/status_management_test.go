//go:build unit

package repository

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRepository_ListWorktrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})

	// Mock repository name extraction
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock status manager to return repository with worktrees
	expectedRepo := &status.Repository{
		Path: "/path/to/repo",
		Remotes: map[string]status.Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
		Worktrees: map[string]status.WorktreeInfo{
			"feature/test-branch": {
				Remote: "origin",
				Branch: "feature/test-branch",
			},
			"bugfix/issue-123": {
				Remote: "origin",
				Branch: "bugfix/issue-123",
			},
		},
	}
	mockStatus.EXPECT().GetRepository("github.com/lerenn/example").Return(expectedRepo, nil)

	result, err := repo.ListWorktrees()
	assert.NoError(t, err)
	assert.Len(t, result, 2, "Should only return worktrees for current repository")

	// Verify specific branches are present
	branchNames := make([]string, len(result))
	for i, wt := range result {
		branchNames[i] = wt.Branch
	}
	assert.Contains(t, branchNames, "feature/test-branch")
	assert.Contains(t, branchNames, "bugfix/issue-123")
	assert.NotContains(t, branchNames, "feature/other-branch")
}

func TestRepository_ListWorktrees_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})

	// Mock repository name extraction
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock status manager to return repository not found
	mockStatus.EXPECT().GetRepository("github.com/lerenn/example").Return(nil, status.ErrRepositoryNotFound)

	result, err := repo.ListWorktrees()
	assert.NoError(t, err)
	assert.Len(t, result, 0, "Should return empty list when repository not found")
}

func TestRepository_AddWorktreeToStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})

	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	err := repo.AddWorktreeToStatus(StatusParams{
		RepoURL:       "github.com/lerenn/example",
		Branch:        "test-branch",
		WorktreePath:  "/test/path",
		WorkspacePath: "",
		Remote:        "origin",
		IssueInfo:     nil,
	})
	assert.NoError(t, err)
}

func TestRepository_RemoveWorktreeFromStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})

	mockWorktree.EXPECT().RemoveFromStatus("github.com/lerenn/example", "test-branch").Return(nil)

	err := repo.RemoveWorktreeFromStatus("github.com/lerenn/example", "test-branch")
	assert.NoError(t, err)
}

func TestRepository_AutoAddRepositoryToStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})

	mockFS.EXPECT().Exists("/test/path/.git").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL("/test/path", "origin").Return("https://github.com/lerenn/example.git", nil)
	mockStatus.EXPECT().AddRepository("github.com/lerenn/example", gomock.Any()).Return(nil)

	err := repo.AutoAddRepositoryToStatus("github.com/lerenn/example", "/test/path")
	assert.NoError(t, err)
}

func TestRepository_AutoAddRepositoryToStatus_NoGitDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})

	mockFS.EXPECT().Exists("/test/path/.git").Return(false, nil)

	err := repo.AutoAddRepositoryToStatus("github.com/lerenn/example", "/test/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a Git repository")
}
