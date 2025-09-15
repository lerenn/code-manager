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

func TestWorktree_AddToStatus_Success(t *testing.T) {
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
