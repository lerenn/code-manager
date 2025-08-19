//go:build unit

package repository

import (
	"fmt"
	"testing"

	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/logger"
	"github.com/lerenn/cm/pkg/prompt"
	"github.com/lerenn/cm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration.
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}
}

func TestRepository_Validate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

	// Mock repository validation - .git exists but git status fails
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("", fmt.Errorf("git error"))

	err := repo.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "git repository is in an invalid state")
}

func TestRepository_Validate_NotClean(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

	// Mock repository validation - .git exists but repository is not clean
	mockFS.EXPECT().Exists(".git").Return(true, nil)
	mockFS.EXPECT().IsDir(".git").Return(true, nil)
	mockGit.EXPECT().Status(".").Return("On branch main\nChanges not staged for commit", nil).AnyTimes()

	err := repo.Validate()
	assert.NoError(t, err)
}

func TestRepository_CreateWorktree_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)
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
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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

func TestRepository_ListWorktrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

	// Mock repository name extraction
	mockGit.EXPECT().GetRepositoryName(".").Return("github.com/lerenn/example", nil)

	// Mock status manager to return worktrees
	allWorktrees := []status.Repository{
		{
			URL:    "github.com/lerenn/example",
			Branch: "feature/test-branch",
			Path:   "/test/base/path/worktrees/github.com/lerenn/example/feature/test-branch",
		},
		{
			URL:    "github.com/other/repo",
			Branch: "feature/other-branch",
			Path:   "/test/base/path/github.com/other/repo/feature/other-branch",
		},
		{
			URL:    "github.com/lerenn/example",
			Branch: "bugfix/issue-123",
			Path:   "/test/base/path/worktrees/github.com/lerenn/example/bugfix/issue-123",
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

func TestRepository_LoadWorktree_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	repo := NewRepository(NewRepositoryParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

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
	mockGit.EXPECT().BranchExistsOnRemote(gomock.Any()).Return(true, nil)

	// Mock worktree creation (reusing existing logic) - called by CreateWorktree
	mockGit.EXPECT().GetRepositoryName(gomock.Any()).Return("github.com/lerenn/example", nil)
	mockStatus.EXPECT().GetWorktree("github.com/lerenn/example", "feature-branch").Return(nil, status.ErrWorktreeNotFound)
	mockGit.EXPECT().IsClean(gomock.Any()).Return(true, nil)
	mockFS.EXPECT().Exists(gomock.Any()).Return(false, nil).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)
	mockGit.EXPECT().BranchExists(gomock.Any(), "feature-branch").Return(false, nil)
	mockGit.EXPECT().CreateBranch(gomock.Any(), "feature-branch").Return(nil)
	mockGit.EXPECT().CreateWorktree(gomock.Any(), gomock.Any(), "feature-branch").Return(nil)

	err := repo.LoadWorktree("origin", "feature-branch")
	assert.NoError(t, err)
}
