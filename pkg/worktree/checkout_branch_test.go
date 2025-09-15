//go:build unit

package worktree

import (
	"errors"
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWorktree_CheckoutBranch_Success(t *testing.T) {
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

	worktreePath := "/test/worktree/path"
	branch := "feature-branch"

	// Mock expectations
	mockGit.EXPECT().CheckoutBranch(worktreePath, branch).Return(nil)

	err := worktree.CheckoutBranch(worktreePath, branch)
	assert.NoError(t, err)
}

func TestWorktree_CheckoutBranch_Error(t *testing.T) {
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

	worktreePath := "/test/worktree/path"
	branch := "feature-branch"
	expectedError := errors.New("checkout failed")

	// Mock expectations
	mockGit.EXPECT().CheckoutBranch(worktreePath, branch).Return(expectedError)

	err := worktree.CheckoutBranch(worktreePath, branch)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to checkout branch in worktree")
}
