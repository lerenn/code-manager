//go:build unit

package worktree

import (
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWorktree_RemoveFromStatus_Success(t *testing.T) {
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

	repoURL := "github.com/octocat/Hello-World"
	branch := "feature-branch"

	// Mock expectations
	mockStatus.EXPECT().RemoveWorktree(repoURL, branch).Return(nil)

	err := worktree.RemoveFromStatus(repoURL, branch)
	assert.NoError(t, err)
}
