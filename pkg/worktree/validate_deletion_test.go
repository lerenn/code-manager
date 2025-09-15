//go:build unit

package worktree

import (
	"errors"
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWorktree_ValidateDeletion_Success(t *testing.T) {
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

func TestWorktree_ValidateDeletion_WorktreeNotInStatus(t *testing.T) {
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

	params := ValidateDeletionParams{
		RepoURL: "github.com/octocat/Hello-World",
		Branch:  "feature-branch",
	}

	// Mock expectations
	mockStatus.EXPECT().GetWorktree(params.RepoURL, params.Branch).Return(nil, errors.New("not found"))

	err := worktree.ValidateDeletion(params)
	assert.ErrorIs(t, err, ErrWorktreeNotInStatus)
}
