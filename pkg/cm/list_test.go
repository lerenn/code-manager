//go:build unit

package cm

import (
	"testing"

	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/git"
	"github.com/lerenn/cm/pkg/ide"
	"github.com/lerenn/cm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCM_ListWorktrees_NoRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockIDE := ide.NewMockManagerInterface(ctrl)

	cm := NewCM(createTestConfig())

	// Override adapters with mocks
	c := cm.(*realCM)
	c.FS = mockFS
	c.Git = mockGit
	c.StatusManager = mockStatus
	c.ideManager = mockIDE

	// Mock single repo detection - no .git found
	mockFS.EXPECT().Exists(".git").Return(false, nil)

	// Mock workspace detection - no workspace files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	result, _, err := cm.ListWorktrees()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no Git repository or workspace found")
	assert.Nil(t, result)
}
