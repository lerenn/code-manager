//go:build unit

package worktree

import (
	"errors"
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWorktree_Delete_Success(t *testing.T) {
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
	mockGit.EXPECT().RemoveWorktree(params.RepoPath, params.WorktreePath, params.Force).Return(nil)
	mockFS.EXPECT().RemoveAll(params.WorktreePath).Return(nil)
	mockStatus.EXPECT().RemoveWorktree(params.RepoURL, params.Branch).Return(nil)

	err := worktree.Delete(params)
	assert.NoError(t, err)
}

func TestWorktree_Delete_WorktreeNotInStatus(t *testing.T) {
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
	mockGit.EXPECT().RemoveWorktree(params.RepoPath, params.WorktreePath, params.Force).Return(nil)
	mockFS.EXPECT().RemoveAll(params.WorktreePath).Return(nil)
	mockStatus.EXPECT().RemoveWorktree(params.RepoURL, params.Branch).Return(nil)

	err := worktree.Delete(params)
	assert.NoError(t, err)
}

func TestWorktree_Delete_ConfirmationCancelled(t *testing.T) {
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
