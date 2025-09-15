//go:build unit

package worktree

import (
	"errors"
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	"github.com/lerenn/code-manager/pkg/git"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWorktree_EnsureBranchExists_BranchExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	worktree := &realWorktree{
		fs:              mockFS,
		git:             mockGit,
		statusManager:   mockStatus,
		prompt:          mockPrompt,
		repositoriesDir: "/test/base",
	}

	repoPath := "/test/repo"
	branch := "feature-branch"

	// Mock expectations
	mockGit.EXPECT().CheckReferenceConflict(repoPath, branch).Return(nil)
	mockGit.EXPECT().BranchExists(repoPath, branch).Return(true, nil)

	err := worktree.EnsureBranchExists(repoPath, branch)
	assert.NoError(t, err)
}

func TestWorktree_EnsureBranchExists_BranchDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	worktree := &realWorktree{
		fs:              mockFS,
		git:             mockGit,
		statusManager:   mockStatus,
		logger:          logger.NewNoopLogger(),
		prompt:          mockPrompt,
		repositoriesDir: "/test/base",
	}

	repoPath := "/test/repo"
	branch := "feature-branch"

	// Mock expectations
	mockGit.EXPECT().CheckReferenceConflict(repoPath, branch).Return(nil)
	mockGit.EXPECT().BranchExists(repoPath, branch).Return(false, nil)
	mockGit.EXPECT().FetchRemote(repoPath, "origin").Return(nil)
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   repoPath,
		RemoteName: "origin",
		Branch:     branch,
	}).Return(false, nil)
	mockGit.EXPECT().GetRemoteURL(repoPath, "origin").Return("https://github.com/octocat/Hello-World.git", nil)
	mockGit.EXPECT().GetDefaultBranch("https://github.com/octocat/Hello-World.git").Return("main", nil)
	mockGit.EXPECT().CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   repoPath,
		NewBranch:  branch,
		FromBranch: "origin/main",
	}).Return(nil)

	err := worktree.EnsureBranchExists(repoPath, branch)
	assert.NoError(t, err)
}

func TestWorktree_EnsureBranchExists_BranchExistsOnRemote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	worktree := &realWorktree{
		fs:              mockFS,
		git:             mockGit,
		statusManager:   mockStatus,
		logger:          logger.NewNoopLogger(),
		prompt:          mockPrompt,
		repositoriesDir: "/test/base",
	}

	repoPath := "/test/repo"
	branch := "feature-branch"

	// Mock expectations
	mockGit.EXPECT().CheckReferenceConflict(repoPath, branch).Return(nil)
	mockGit.EXPECT().BranchExists(repoPath, branch).Return(false, nil)
	mockGit.EXPECT().FetchRemote(repoPath, "origin").Return(nil)
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   repoPath,
		RemoteName: "origin",
		Branch:     branch,
	}).Return(true, nil)
	mockGit.EXPECT().CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   repoPath,
		NewBranch:  branch,
		FromBranch: "origin/" + branch,
	}).Return(nil)

	err := worktree.EnsureBranchExists(repoPath, branch)
	assert.NoError(t, err)
}

func TestWorktree_EnsureBranchExists_BranchDoesNotExist_RemoteFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	worktree := &realWorktree{
		fs:              mockFS,
		git:             mockGit,
		statusManager:   mockStatus,
		logger:          logger.NewNoopLogger(),
		prompt:          mockPrompt,
		repositoriesDir: "/test/base",
	}

	repoPath := "/test/repo"
	branch := "feature-branch"

	// Mock expectations - remote operations fail, fallback to local
	mockGit.EXPECT().CheckReferenceConflict(repoPath, branch).Return(nil)
	mockGit.EXPECT().BranchExists(repoPath, branch).Return(false, nil)
	mockGit.EXPECT().FetchRemote(repoPath, "origin").Return(nil)
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   repoPath,
		RemoteName: "origin",
		Branch:     branch,
	}).Return(false, nil)
	mockGit.EXPECT().GetRemoteURL(repoPath, "origin").Return("", errors.New("no remote"))
	mockGit.EXPECT().GetCurrentBranch(repoPath).Return("main", nil)
	mockGit.EXPECT().CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   repoPath,
		NewBranch:  branch,
		FromBranch: "main",
	}).Return(nil)

	err := worktree.EnsureBranchExists(repoPath, branch)
	assert.NoError(t, err)
}

func TestWorktree_EnsureBranchExists_BranchDoesNotExist_DefaultBranchFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	worktree := &realWorktree{
		fs:              mockFS,
		git:             mockGit,
		statusManager:   mockStatus,
		logger:          logger.NewNoopLogger(),
		prompt:          mockPrompt,
		repositoriesDir: "/test/base",
	}

	repoPath := "/test/repo"
	branch := "feature-branch"

	// Mock expectations - remote exists but default branch detection fails, fallback to local
	mockGit.EXPECT().CheckReferenceConflict(repoPath, branch).Return(nil)
	mockGit.EXPECT().BranchExists(repoPath, branch).Return(false, nil)
	mockGit.EXPECT().FetchRemote(repoPath, "origin").Return(nil)
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   repoPath,
		RemoteName: "origin",
		Branch:     branch,
	}).Return(false, nil)
	mockGit.EXPECT().GetRemoteURL(repoPath, "origin").Return("https://github.com/octocat/Hello-World.git", nil)
	mockGit.EXPECT().GetDefaultBranch("https://github.com/octocat/Hello-World.git").Return("", errors.New("failed to get default branch"))
	mockGit.EXPECT().GetCurrentBranch(repoPath).Return("main", nil)
	mockGit.EXPECT().CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   repoPath,
		NewBranch:  branch,
		FromBranch: "main",
	}).Return(nil)

	err := worktree.EnsureBranchExists(repoPath, branch)
	assert.NoError(t, err)
}
