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
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWorktree_Create_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	worktree := &realWorktree{
		fs: mockFS, git: mockGit, statusManager: mockStatus, logger: logger.NewNoopLogger(), prompt: mockPrompt,
		repositoriesDir: "/test/base",
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
	mockGit.EXPECT().CheckReferenceConflict(params.RepoPath, params.Branch).Return(nil)
	mockGit.EXPECT().BranchExists(params.RepoPath, params.Branch).Return(true, nil)
	mockFS.EXPECT().MkdirAll(params.WorktreePath, gomock.Any()).Return(nil)
	mockGit.EXPECT().CreateWorktreeWithNoCheckout(params.RepoPath, params.WorktreePath, params.Branch).Return(nil)

	err := worktree.Create(params)
	assert.NoError(t, err)
}

func TestWorktree_Create_DirectoryExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	worktree := &realWorktree{
		fs: mockFS, git: mockGit, statusManager: mockStatus, prompt: mockPrompt, repositoriesDir: "/test/base",
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
	mockGit.EXPECT().CheckReferenceConflict(params.RepoPath, params.Branch).Return(nil)
	mockGit.EXPECT().BranchExists(params.RepoPath, params.Branch).Return(false, nil)
	mockGit.EXPECT().FetchRemote(params.RepoPath, "origin").Return(nil)
	mockGit.EXPECT().BranchExistsOnRemote(git.BranchExistsOnRemoteParams{
		RepoPath:   params.RepoPath,
		RemoteName: "origin",
		Branch:     params.Branch,
	}).Return(false, nil)
	mockGit.EXPECT().GetRemoteURL(params.RepoPath, "origin").Return("https://github.com/octocat/Hello-World.git", nil)
	mockGit.EXPECT().GetDefaultBranch("https://github.com/octocat/Hello-World.git").Return("main", nil)
	mockGit.EXPECT().CreateBranchFrom(git.CreateBranchFromParams{
		RepoPath:   params.RepoPath,
		NewBranch:  params.Branch,
		FromBranch: "origin/main",
	}).Return(nil)
	mockFS.EXPECT().MkdirAll(params.WorktreePath, gomock.Any()).Return(nil)
	mockGit.EXPECT().CreateWorktreeWithNoCheckout(params.RepoPath, params.WorktreePath, params.Branch).Return(nil)

	err := worktree.Create(params)
	assert.NoError(t, err)
}
