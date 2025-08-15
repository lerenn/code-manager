//go:build unit

package wtm

import (
	"testing"

	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/logger"
	"github.com/lerenn/wtm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestRepository_Validate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil)
	mockGit.EXPECT().Status(".").Return("On branch main", nil) // Called twice for validation

	err := repo.Validate()
	assert.NoError(t, err)
}

func TestRepository_Validate_NoGitDir(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation - .git not found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	err := repo.Validate()
	assert.ErrorIs(t, err, ErrGitRepositoryNotFound)
}

func TestRepository_Validate_GitStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("", assert.AnError)

	err := repo.Validate()
	assert.ErrorIs(t, err, ErrGitRepositoryInvalid)
}

func TestRepository_CreateWorktree_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree creation
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "test-branch", gomock.Any(), "").Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "test-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "test-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "test-branch").Return(nil)

	err := repo.CreateWorktree("test-branch")
	assert.NoError(t, err)
}

func TestRepository_IsWorkspaceFile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock workspace files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	exists, err := repo.IsWorkspaceFile()
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestRepository_IsWorkspaceFile_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock no workspace files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	exists, err := repo.IsWorkspaceFile()
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRepository_IsGitRepository_Directory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock .git exists and is a directory (regular repository)
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	exists, err := repo.IsGitRepository()
	assert.NoError(t, err)
	assert.True(t, exists) // Should return true for regular repositories
}

func TestRepository_IsGitRepository_File(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock .git exists but is not a directory (worktree case)
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(false, nil)
	// Mock .git file content with valid worktree format
	mockFS.EXPECT().ReadFile(".git").Return([]byte("gitdir: /path/to/main/repo/.git/worktrees/worktree-name"), nil)

	exists, err := repo.IsGitRepository()
	assert.NoError(t, err)
	assert.True(t, exists) // Should return true for valid worktrees
}

func TestRepository_IsGitRepository_InvalidFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock .git exists but is not a directory
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(false, nil)
	// Mock .git file content without valid worktree format
	mockFS.EXPECT().ReadFile(".git").Return([]byte("not a git worktree file"), nil)

	exists, err := repo.IsGitRepository()
	assert.NoError(t, err)
	assert.False(t, exists) // Should return false for invalid .git files
}

func TestRepository_IsGitRepository_NotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock .git does not exist
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	exists, err := repo.IsGitRepository()
	assert.NoError(t, err)
	assert.False(t, exists) // Should return false when .git doesn't exist
}

func TestRepository_DeleteWorktree_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree deletion
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(&status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "test-branch",
		Path:   "/test/path/worktree",
	}, nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "test-branch").Return("/test/path/worktree", nil)
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), "/test/path/worktree").Return(nil)
	mockFS.EXPECT().RemoveAll("/test/path/worktree").Return(nil)
	mockStatus.EXPECT().RemoveWorktree("github.com/lerenn/example", "test-branch").Return(nil)

	err := repo.DeleteWorktree("test-branch", true) // Force deletion
	assert.NoError(t, err)
}

func TestRepository_DeleteWorktree_NotInStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree not found in status
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(nil, status.ErrWorktreeNotFound)

	err := repo.DeleteWorktree("test-branch", true)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
}

func TestRepository_DeleteWorktree_GetWorktreePathError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree found in status but Git path lookup fails
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(&status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "test-branch",
		Path:   "/test/path/worktree",
	}, nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "test-branch").Return("", assert.AnError)

	err := repo.DeleteWorktree("test-branch", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get worktree path")
}

func TestRepository_DeleteWorktree_RemoveWorktreeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree deletion with Git removal error
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(&status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "test-branch",
		Path:   "/test/path/worktree",
	}, nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "test-branch").Return("/test/path/worktree", nil)
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), "/test/path/worktree").Return(assert.AnError)

	err := repo.DeleteWorktree("test-branch", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove worktree from Git")
}

func TestRepository_DeleteWorktree_RemoveAllError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree deletion with file system removal error
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(&status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "test-branch",
		Path:   "/test/path/worktree",
	}, nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "test-branch").Return("/test/path/worktree", nil)
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), "/test/path/worktree").Return(nil)
	mockFS.EXPECT().RemoveAll("/test/path/worktree").Return(assert.AnError)

	err := repo.DeleteWorktree("test-branch", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove worktree directory")
}

func TestRepository_DeleteWorktree_StatusRemoveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock worktree deletion with status removal error
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "test-branch").Return(&status.Repository{
		URL:    "github.com/lerenn/example",
		Branch: "test-branch",
		Path:   "/test/path/worktree",
	}, nil)
	mockGit.EXPECT().GetWorktreePath(gomock.Any(), "test-branch").Return("/test/path/worktree", nil)
	mockGit.EXPECT().RemoveWorktree(gomock.Any(), "/test/path/worktree").Return(nil)
	mockFS.EXPECT().RemoveAll("/test/path/worktree").Return(nil)
	mockStatus.EXPECT().RemoveWorktree("github.com/lerenn/example", "test-branch").Return(assert.AnError)

	err := repo.DeleteWorktree("test-branch", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove worktree from status")
}

func TestRepository_ListWorktrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository name extraction
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock status manager to return worktrees
	allWorktrees := []status.Repository{
		{
			URL:    "github.com/lerenn/example",
			Branch: "feature/test-branch",
			Path:   "/test/base/path/github.com/lerenn/example/feature/test-branch",
		},
		{
			URL:    "github.com/other/repo",
			Branch: "feature/other-branch",
			Path:   "/test/base/path/github.com/other/repo/feature/other-branch",
		},
		{
			URL:    "github.com/lerenn/example",
			Branch: "bugfix/issue-123",
			Path:   "/test/base/path/github.com/lerenn/example/bugfix/issue-123",
		},
	}
	mockStatus.EXPECT().ListAllWorktrees().Return(allWorktrees, nil)

	// Mock GetBranchRemote calls for the filtered worktrees
	mockGit.EXPECT().GetBranchRemote(".", "feature/test-branch").Return("origin", nil)
	mockGit.EXPECT().GetBranchRemote(".", "bugfix/issue-123").Return("origin", nil)

	result, err := repo.ListWorktrees()
	assert.NoError(t, err)
	assert.Len(t, result, 2, "Should only return worktrees for current repository")

	// Verify only current repository worktrees are returned
	for _, wt := range result {
		assert.Equal(t, "github.com/lerenn/example", wt.URL)
	}

	// Verify specific branches are present
	branchNames := make([]string, len(result))
	for i, wt := range result {
		branchNames[i] = wt.Branch
	}
	assert.Contains(t, branchNames, "feature/test-branch")
	assert.Contains(t, branchNames, "bugfix/issue-123")
	assert.NotContains(t, branchNames, "feature/other-branch")
}

func TestRepository_ListWorktrees_NoWorktrees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository name extraction
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock status manager to return empty list
	mockStatus.EXPECT().ListAllWorktrees().Return([]status.Repository{}, nil)

	result, err := repo.ListWorktrees()
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

func TestRepository_ListWorktrees_RepositoryNameError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository name extraction to fail
	mockGit.EXPECT().GetRepositoryName(".").Return("", assert.AnError)

	result, err := repo.ListWorktrees()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get repository name")
	assert.Nil(t, result)
}

func TestRepository_ListWorktrees_StatusFileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository name extraction
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock status manager to return error
	mockStatus.EXPECT().ListAllWorktrees().Return(nil, assert.AnError)

	result, err := repo.ListWorktrees()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load worktrees from status file")
	assert.Nil(t, result)
}

func TestRepository_ListWorktrees_Filtering(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock repository name extraction
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock status manager to return worktrees from multiple repositories
	allWorktrees := []status.Repository{
		{
			URL:    "github.com/lerenn/example",
			Branch: "feature/test-branch",
			Path:   "/test/base/path/github.com/lerenn/example/feature/test-branch",
		},
		{
			URL:    "github.com/other/repo",
			Branch: "feature/other-branch",
			Path:   "/test/base/path/github.com/other/repo/feature/other-branch",
		},
		{
			URL:    "github.com/another/repo",
			Branch: "feature/another-branch",
			Path:   "/test/base/path/github.com/another/repo/feature/another-branch",
		},
		{
			URL:    "github.com/lerenn/example",
			Branch: "bugfix/issue-123",
			Path:   "/test/base/path/github.com/lerenn/example/bugfix/issue-123",
		},
	}
	mockStatus.EXPECT().ListAllWorktrees().Return(allWorktrees, nil)

	// Mock GetBranchRemote calls for the filtered worktrees
	mockGit.EXPECT().GetBranchRemote(".", "feature/test-branch").Return("origin", nil)
	mockGit.EXPECT().GetBranchRemote(".", "bugfix/issue-123").Return("origin", nil)

	result, err := repo.ListWorktrees()
	assert.NoError(t, err)
	assert.Len(t, result, 2, "Should only return worktrees for current repository")

	// Verify only current repository worktrees are returned
	for _, wt := range result {
		assert.Equal(t, "github.com/lerenn/example", wt.URL)
	}

	// Verify specific branches are present
	branchNames := make([]string, len(result))
	for i, wt := range result {
		branchNames[i] = wt.Branch
	}
	assert.Contains(t, branchNames, "feature/test-branch")
	assert.Contains(t, branchNames, "bugfix/issue-123")
	assert.NotContains(t, branchNames, "feature/other-branch")
	assert.NotContains(t, branchNames, "feature/another-branch")
}

func TestRepository_LoadWorktree_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation (called by LoadWorktree and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "origin").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(".", "origin", "feature-branch").Return(true, nil)

	// Mock worktree creation (reusing existing logic) - called by CreateWorktree
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "feature-branch", gomock.Any(), "").Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	err := repo.LoadWorktree("origin", "feature-branch")
	assert.NoError(t, err)
}

func TestRepository_LoadBranch_NewRemote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation (called by LoadBranch and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock remote management (new remote)
	mockGit.EXPECT().RemoteExists(".", "otheruser").Return(false, nil)
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)
	mockGit.EXPECT().AddRemote(".", "otheruser", "https://github.com/otheruser/example.git").Return(nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "otheruser").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(".", "otheruser", "feature-branch").Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "feature-branch", gomock.Any(), "").Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	err := repo.LoadWorktree("otheruser", "feature-branch")
	assert.NoError(t, err)
}

func TestRepository_LoadWorktree_SSHProtocol(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation (called by LoadWorktree and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation (SSH)
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("git@github.com:lerenn/example.git", nil)

	// Mock remote management (new remote with SSH)
	mockGit.EXPECT().RemoteExists(".", "otheruser").Return(false, nil)
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("git@github.com:lerenn/example.git", nil)
	mockGit.EXPECT().AddRemote(".", "otheruser", "git@github.com:otheruser/example.git").Return(nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "otheruser").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(".", "otheruser", "feature-branch").Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "feature-branch", gomock.Any(), "").Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	err := repo.LoadWorktree("otheruser", "feature-branch")
	assert.NoError(t, err)
}

// TestRepository_URLConstruction tests the URL construction logic for different protocols
func TestRepository_URLConstruction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	tests := []struct {
		name         string
		originURL    string
		remoteSource string
		repoName     string
		expectedURL  string
	}{
		{
			name:         "HTTPS protocol",
			originURL:    "https://github.com/lerenn/example.git",
			remoteSource: "otheruser",
			repoName:     "github.com/lerenn/example",
			expectedURL:  "https://github.com/otheruser/example.git",
		},
		{
			name:         "SSH protocol",
			originURL:    "git@github.com:lerenn/example.git",
			remoteSource: "otheruser",
			repoName:     "github.com/lerenn/example",
			expectedURL:  "git@github.com:otheruser/example.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock Git repository validation
			mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
			mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()
			mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

			// Mock origin remote validation
			mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
			mockGit.EXPECT().GetRemoteURL(".", "origin").Return(tt.originURL, nil)

			// Mock remote management
			mockGit.EXPECT().RemoteExists(".", tt.remoteSource).Return(false, nil)
			mockGit.EXPECT().GetRepositoryName(".").Return(tt.repoName, nil)
			mockGit.EXPECT().GetRemoteURL(".", "origin").Return(tt.originURL, nil)
			mockGit.EXPECT().AddRemote(".", tt.remoteSource, tt.expectedURL).Return(nil)

			// Mock fetch and branch check
			mockGit.EXPECT().FetchRemote(".", tt.remoteSource).Return(nil)
			mockGit.EXPECT().BranchExistsOnRemote(".", tt.remoteSource, "feature-branch").Return(true, nil)

			// Mock worktree creation
			mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return(tt.repoName, nil)
			mockStatus.EXPECT().GetWorktree(tt.repoName, "feature-branch").Return(nil, status.ErrWorktreeNotFound)
			mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
			mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
			mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
			mockStatus.EXPECT().AddWorktree(tt.repoName, "feature-branch", gomock.Any(), "").Return(nil)
			mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
			mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
			mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

			err := repo.LoadWorktree(tt.remoteSource, "feature-branch")
			assert.NoError(t, err)
		})
	}
}

func TestRepository_LoadWorktree_DefaultRemote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation (called by LoadWorktree and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "origin").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(".", "origin", "feature-branch").Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "feature-branch", gomock.Any(), "").Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	// Test with empty remote source (should default to "origin")
	err := repo.LoadWorktree("", "feature-branch")
	assert.NoError(t, err)
}

func TestRepository_LoadWorktree_GitRepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository not found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	err := repo.LoadWorktree("origin", "feature-branch")
	assert.ErrorIs(t, err, ErrGitRepositoryNotFound)
}

func TestRepository_LoadWorktree_OriginRemoteNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	// Mock origin remote not found
	mockGit.EXPECT().RemoteExists(".", "origin").Return(false, nil)

	err := repo.LoadWorktree("origin", "feature-branch")
	assert.ErrorIs(t, err, ErrOriginRemoteNotFound)
}

func TestRepository_LoadWorktree_OriginRemoteInvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)

	// Mock origin remote exists but invalid URL
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("invalid-url-format", nil)

	err := repo.LoadWorktree("origin", "feature-branch")
	assert.ErrorIs(t, err, ErrOriginRemoteInvalidURL)
}

func TestRepository_LoadWorktree_FetchFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation (called by LoadWorktree and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote fails
	mockGit.EXPECT().FetchRemote(".", "origin").Return(assert.AnError)

	err := repo.LoadWorktree("origin", "feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch from remote")
}

func TestRepository_LoadWorktree_BranchNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation (called by LoadWorktree and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "origin").Return(nil)

	// Mock branch existence check fails
	mockGit.EXPECT().BranchExistsOnRemote(".", "origin", "feature-branch").Return(false, nil)

	err := repo.LoadWorktree("origin", "feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "branch not found on remote")
}

func TestRepository_LoadWorktree_AddRemoteFailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation (called by LoadBranch and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock remote management (new remote fails to add)
	mockGit.EXPECT().RemoteExists(".", "otheruser").Return(false, nil)
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)
	mockGit.EXPECT().AddRemote(".", "otheruser", "https://github.com/otheruser/example.git").Return(assert.AnError)

	err := repo.LoadWorktree("otheruser", "feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add remote")
}

func TestRepository_LoadWorktree_ExistingRemote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation (called by LoadWorktree and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock remote management (existing remote)
	mockGit.EXPECT().RemoteExists(".", "otheruser").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "otheruser").Return("https://github.com/otheruser/example.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "otheruser").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(".", "otheruser", "feature-branch").Return(true, nil)

	// Mock worktree creation (reusing existing logic)
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "feature-branch", gomock.Any(), "").Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	err := repo.LoadWorktree("otheruser", "feature-branch")
	assert.NoError(t, err)
}

func TestRepository_LoadWorktree_WorktreeCreationError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Mock Git repository validation (called by LoadWorktree and CreateWorktree)
	mockFS.EXPECT().Exists(".git").Return(true, nil).AnyTimes()
	mockFS.EXPECT().IsDir(".git").Return(true, nil).AnyTimes()

	// Mock Git status for validation (called by CreateWorktree)
	mockGit.EXPECT().Status(".").Return("On branch main", nil).AnyTimes()

	// Mock origin remote validation
	mockGit.EXPECT().RemoteExists(".", "origin").Return(true, nil)
	mockGit.EXPECT().GetRemoteURL(".", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock fetch from remote
	mockGit.EXPECT().FetchRemote(".", "origin").Return(nil)

	// Mock branch existence check
	mockGit.EXPECT().BranchExistsOnRemote(".", "origin", "feature-branch").Return(true, nil)

	// Mock worktree creation fails
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree("github.com/lerenn/example", "feature-branch", gomock.Any(), "").Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(assert.AnError)
	// Mock cleanup calls
	mockStatus.EXPECT().RemoveWorktree("github.com/lerenn/example", "feature-branch").Return(nil)

	err := repo.LoadWorktree("origin", "feature-branch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create Git worktree")
}

func TestRepository_ExtractHostFromURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Test cases for different URL formats
	testCases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "GitHub HTTPS URL",
			url:      "https://github.com/lerenn/example.git",
			expected: "github.com",
		},
		{
			name:     "GitHub SSH URL",
			url:      "git@github.com:lerenn/example.git",
			expected: "github.com",
		},
		{
			name:     "GitLab HTTPS URL",
			url:      "https://gitlab.com/lerenn/example.git",
			expected: "gitlab.com",
		},
		{
			name:     "GitLab SSH URL",
			url:      "git@gitlab.com:lerenn/example.git",
			expected: "gitlab.com",
		},
		{
			name:     "Bitbucket HTTPS URL",
			url:      "https://bitbucket.org/lerenn/example.git",
			expected: "bitbucket.org",
		},
		{
			name:     "Bitbucket SSH URL",
			url:      "git@bitbucket.org:lerenn/example.git",
			expected: "bitbucket.org",
		},
		{
			name:     "Custom Git server HTTPS URL",
			url:      "https://git.company.com/lerenn/example.git",
			expected: "git.company.com",
		},
		{
			name:     "Custom Git server SSH URL",
			url:      "git@git.company.com:lerenn/example.git",
			expected: "git.company.com",
		},
		{
			name:     "URL without .git suffix",
			url:      "https://github.com/lerenn/example",
			expected: "github.com",
		},
		{
			name:     "Invalid URL format",
			url:      "invalid-url-format",
			expected: "",
		},
		{
			name:     "Empty URL",
			url:      "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := repo.extractHostFromURL(tc.url)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestRepository_HandleRemoteManagement_DifferentHosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)

	repo := newRepository(mockFS, mockGit, createTestConfig(), mockStatus, logger.NewNoopLogger(), true)

	// Test cases for different hosts and protocols
	testCases := []struct {
		name          string
		originURL     string
		remoteSource  string
		repoName      string
		expectedSSH   string
		expectedHTTPS string
	}{
		{
			name:          "GitHub repository",
			originURL:     "https://github.com/lerenn/example.git",
			remoteSource:  "upstream",
			repoName:      "github.com/lerenn/example",
			expectedSSH:   "git@github.com:upstream/example.git",
			expectedHTTPS: "https://github.com/upstream/example.git",
		},
		{
			name:          "GitLab repository",
			originURL:     "git@gitlab.com:lerenn/example.git",
			remoteSource:  "fork",
			repoName:      "gitlab.com/lerenn/example",
			expectedSSH:   "git@gitlab.com:fork/example.git",
			expectedHTTPS: "https://gitlab.com/fork/example.git",
		},
		{
			name:          "Bitbucket repository",
			originURL:     "https://bitbucket.org/lerenn/example.git",
			remoteSource:  "upstream",
			repoName:      "bitbucket.org/lerenn/example",
			expectedSSH:   "git@bitbucket.org:upstream/example.git",
			expectedHTTPS: "https://bitbucket.org/upstream/example.git",
		},
		{
			name:          "Custom Git server",
			originURL:     "git@git.company.com:lerenn/example.git",
			remoteSource:  "staging",
			repoName:      "git.company.com/lerenn/example",
			expectedSSH:   "git@git.company.com:staging/example.git",
			expectedHTTPS: "https://git.company.com/staging/example.git",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mock that remote doesn't exist
			mockGit.EXPECT().RemoteExists(".", tc.remoteSource).Return(false, nil)

			// Mock getting repository name
			mockGit.EXPECT().GetRepositoryName(".").Return(tc.repoName, nil)

			// Mock getting origin URL
			mockGit.EXPECT().GetRemoteURL(".", "origin").Return(tc.originURL, nil)

			// Mock adding remote - we'll capture the URL that gets passed
			var capturedURL string
			mockGit.EXPECT().AddRemote(".", tc.remoteSource, gomock.Any()).DoAndReturn(
				func(repoPath, remoteName, remoteURL string) error {
					capturedURL = remoteURL
					return nil
				},
			)

			// Execute the method
			err := repo.handleRemoteManagement(tc.remoteSource)
			assert.NoError(t, err)

			// Verify the constructed URL matches expected format based on protocol
			protocol := repo.determineProtocol(tc.originURL)
			if protocol == "ssh" {
				assert.Equal(t, tc.expectedSSH, capturedURL)
			} else {
				assert.Equal(t, tc.expectedHTTPS, capturedURL)
			}
		})
	}
}
