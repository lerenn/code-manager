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

func TestNewWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)
	mockPrompt := promptmocks.NewMockPrompter(ctrl)

	params := NewWorktreeParams{
		FS:              mockFS,
		Git:             mockGit,
		StatusManager:   mockStatus,
		Prompt:          mockPrompt,
		RepositoriesDir: "/test/path",
	}

	worktree := NewWorktree(params)
	assert.NotNil(t, worktree)
}
