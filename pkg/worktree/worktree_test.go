//go:build unit

package worktree

import (
	"errors"
	"testing"

	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	params := NewWorktreeParams{
		FS:            mockFS,
		Git:           mockGit,
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		BasePath:      "/test/path",
		Verbose:       true,
	}

	worktree := NewWorktree(params)
	assert.NotNil(t, worktree)
}

func TestWorktree_BuildPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	path := worktree.BuildPath("github.com/octocat/Hello-World", "origin", "feature-branch")
	expected := "/test/base/github.com/octocat/Hello-World/origin/feature-branch"
	assert.Equal(t, expected, path)
}

func TestWorktree_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := CreateParams{
		RepoURL:      "github.com/octocat/Hello-World",
		Branch:       "feature-branch",
		WorktreePath: "/test/base/github.com/octocat/Hello-World/origin/feature-branch",
		RepoPath:     "/test/repo",
		Remote:       "origin",
		IssueInfo:    nil,
		Force:        false,
	}

	// Mock expectations
	mockFS.EXPECT().Exists(params.WorktreePath).Return(false, nil)
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(nil, errors.New("not found"))
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockGit.EXPECT().BranchExists(params.RepoPath, params.Branch).Return(true, nil)
	mockFS.EXPECT().MkdirAll(params.WorktreePath, gomock.Any()).Return(nil)
	mockGit.EXPECT().CreateWorktree(params.RepoPath, params.WorktreePath, params.Branch).Return(nil)

	err := worktree.Create(params)
	assert.NoError(t, err)
}

func TestWorktree_Create_DirectoryExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := CreateParams{
		RepoURL:      "github.com/octocat/Hello-World",
		Branch:       "feature-branch",
		WorktreePath: "/test/base/github.com/octocat/Hello-World/origin/feature-branch",
		RepoPath:     "/test/repo",
		Remote:       "origin",
		IssueInfo:    nil,
		Force:        false,
	}

	// Mock expectations
	mockFS.EXPECT().Exists(params.WorktreePath).Return(true, nil)

	err := worktree.Create(params)
	assert.ErrorIs(t, err, ErrDirectoryExists)
}

func TestWorktree_Create_WorktreeExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := CreateParams{
		RepoURL:      "github.com/octocat/Hello-World",
		Branch:       "feature-branch",
		WorktreePath: "/test/base/github.com/octocat/Hello-World/origin/feature-branch",
		RepoPath:     "/test/repo",
		Remote:       "origin",
		IssueInfo:    nil,
		Force:        false,
	}

	existingWorktree := &status.WorktreeInfo{
		Branch: params.Branch,
		Remote: "origin",
	}

	// Mock expectations
	mockFS.EXPECT().Exists(params.WorktreePath).Return(false, nil)
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(existingWorktree, nil)

	err := worktree.Create(params)
	assert.ErrorIs(t, err, ErrWorktreeExists)
}

func TestWorktree_Create_BranchDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := CreateParams{
		RepoURL:      "github.com/octocat/Hello-World",
		Branch:       "feature-branch",
		WorktreePath: "/test/base/github.com/octocat/Hello-World/origin/feature-branch",
		RepoPath:     "/test/repo",
		Remote:       "origin",
		IssueInfo:    nil,
		Force:        false,
	}

	// Mock expectations
	mockFS.EXPECT().Exists(params.WorktreePath).Return(false, nil)
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(nil, errors.New("not found"))
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)
	mockGit.EXPECT().BranchExists(params.RepoPath, params.Branch).Return(false, nil)
	mockGit.EXPECT().CreateBranch(params.RepoPath, params.Branch).Return(nil)
	mockFS.EXPECT().MkdirAll(params.WorktreePath, gomock.Any()).Return(nil)
	mockGit.EXPECT().CreateWorktree(params.RepoPath, params.WorktreePath, params.Branch).Return(nil)

	err := worktree.Create(params)
	assert.NoError(t, err)
}

func TestWorktree_Delete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := DeleteParams{
		RepoURL:      "github.com/octocat/Hello-World",
		Branch:       "feature-branch",
		WorktreePath: "/test/base/github.com/octocat/Hello-World/origin/feature-branch",
		RepoPath:     "/test/repo",
		Force:        true,
	}

	existingWorktree := &status.WorktreeInfo{
		Branch: params.Branch,
		Remote: "origin",
	}

	// Mock expectations
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(existingWorktree, nil)
	mockGit.EXPECT().RemoveWorktree(params.RepoPath, params.WorktreePath).Return(nil)
	mockFS.EXPECT().RemoveAll(params.WorktreePath).Return(nil)
	mockStatus.EXPECT().RemoveWorktree(params.RepoURL, params.Branch).Return(nil)

	err := worktree.Delete(params)
	assert.NoError(t, err)
}

func TestWorktree_Delete_WorktreeNotInStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := DeleteParams{
		RepoURL:      "github.com/octocat/Hello-World",
		Branch:       "feature-branch",
		WorktreePath: "/test/base/github.com/octocat/Hello-World/origin/feature-branch",
		RepoPath:     "/test/repo",
		Force:        true,
	}

	// Mock expectations
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(nil, errors.New("not found"))

	err := worktree.Delete(params)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
}

func TestWorktree_Delete_WithConfirmation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := DeleteParams{
		RepoURL:      "github.com/octocat/Hello-World",
		Branch:       "feature-branch",
		WorktreePath: "/test/base/github.com/octocat/Hello-World/origin/feature-branch",
		RepoPath:     "/test/repo",
		Force:        false,
	}

	existingWorktree := &status.WorktreeInfo{
		Branch: params.Branch,
		Remote: "origin",
	}

	// Mock expectations
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(existingWorktree, nil)
	mockPrompt.EXPECT().PromptForConfirmation(gomock.Any(), false).Return(true, nil)
	mockGit.EXPECT().RemoveWorktree(params.RepoPath, params.WorktreePath).Return(nil)
	mockFS.EXPECT().RemoveAll(params.WorktreePath).Return(nil)
	mockStatus.EXPECT().RemoveWorktree(params.RepoURL, params.Branch).Return(nil)

	err := worktree.Delete(params)
	assert.NoError(t, err)
}

func TestWorktree_Delete_ConfirmationCancelled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := DeleteParams{
		RepoURL:      "github.com/octocat/Hello-World",
		Branch:       "feature-branch",
		WorktreePath: "/test/base/github.com/octocat/Hello-World/origin/feature-branch",
		RepoPath:     "/test/repo",
		Force:        false,
	}

	existingWorktree := &status.WorktreeInfo{
		Branch: params.Branch,
		Remote: "origin",
	}

	// Mock expectations
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(existingWorktree, nil)
	mockPrompt.EXPECT().PromptForConfirmation(gomock.Any(), false).Return(false, nil)

	err := worktree.Delete(params)
	assert.ErrorIs(t, err, ErrDeletionCancelled)
}

func TestWorktree_ValidateCreation_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := ValidateCreationParams{
		RepoURL:      "github.com/octocat/Hello-World",
		Branch:       "feature-branch",
		WorktreePath: "/test/base/github.com/octocat/Hello-World/origin/feature-branch",
		RepoPath:     "/test/repo",
	}

	// Mock expectations
	mockFS.EXPECT().Exists(params.WorktreePath).Return(false, nil)
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(nil, errors.New("not found"))
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil)

	err := worktree.ValidateCreation(params)
	assert.NoError(t, err)
}

func TestWorktree_ValidateDeletion_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := ValidateDeletionParams{
		RepoURL: "github.com/octocat/Hello-World",
		Branch:  "feature-branch",
	}

	existingWorktree := &status.WorktreeInfo{
		Branch: params.Branch,
		Remote: "origin",
	}

	// Mock expectations
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(existingWorktree, nil)

	err := worktree.ValidateDeletion(params)
	assert.NoError(t, err)
}

func TestWorktree_EnsureBranchExists_BranchExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	repoPath := "/test/repo"
	branch := "feature-branch"

	// Mock expectations
	mockGit.EXPECT().BranchExists(repoPath, branch).Return(true, nil)

	err := worktree.EnsureBranchExists(repoPath, branch)
	assert.NoError(t, err)
}

func TestWorktree_EnsureBranchExists_BranchDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	repoPath := "/test/repo"
	branch := "feature-branch"

	// Mock expectations
	mockGit.EXPECT().BranchExists(repoPath, branch).Return(false, nil)
	mockGit.EXPECT().CreateBranch(repoPath, branch).Return(nil)

	err := worktree.EnsureBranchExists(repoPath, branch)
	assert.NoError(t, err)
}

func TestWorktree_AddToStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	params := AddToStatusParams{
		RepoURL:       "github.com/octocat/Hello-World",
		Branch:        "feature-branch",
		WorktreePath:  "/test/path",
		WorkspacePath: "",
		Remote:        "origin",
		IssueInfo:     nil,
	}

	// Mock expectations
	mockStatus.EXPECT().AddWorktree(gomock.Any()).Return(nil)

	err := worktree.AddToStatus(params)
	assert.NoError(t, err)
}

func TestWorktree_RemoveFromStatus_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	repoURL := "github.com/octocat/Hello-World"
	branch := "feature-branch"

	// Mock expectations
	mockStatus.EXPECT().RemoveWorktree(repoURL, branch).Return(nil)

	err := worktree.RemoveFromStatus(repoURL, branch)
	assert.NoError(t, err)
}

func TestWorktree_Exists_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	repoPath := "/test/repo"
	branch := "feature-branch"

	// Mock expectations
	mockGit.EXPECT().WorktreeExists(repoPath, branch).Return(true, nil)

	exists, err := worktree.Exists(repoPath, branch)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestWorktree_GetPath_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	repoPath := "/test/repo"
	branch := "feature-branch"
	expectedPath := "/test/worktree/path"

	// Mock expectations
	mockGit.EXPECT().GetWorktreePath(repoPath, branch).Return(expectedPath, nil)

	path, err := worktree.GetPath(repoPath, branch)
	assert.NoError(t, err)
	assert.Equal(t, expectedPath, path)
}

func TestWorktree_CleanupDirectory_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockPrompt := prompt.NewMockPrompt(ctrl)

	worktree := &realWorktree{
		fs:            mockFS,
		git:           mockGit,
		statusManager: mockStatus,
		logger:        mockLogger,
		prompt:        mockPrompt,
		basePath:      "/test/base",
		verbose:       false,
	}

	worktreePath := "/test/worktree/path"

	// Mock expectations
	mockFS.EXPECT().RemoveAll(worktreePath).Return(nil)

	err := worktree.CleanupDirectory(worktreePath)
	assert.NoError(t, err)
}
