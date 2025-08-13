//go:build unit

package wtm

import (
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/git"
	"github.com/lerenn/wtm/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createStatusTestConfig creates a test configuration for status tests.
func createStatusTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtm/status.yaml",
	}
}

func TestAddWorktreeToStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	wtm := NewWTM(createStatusTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	worktreePath := "/home/user/.wtm/repos/github.com/lerenn/example/feature-a"
	workspacePath := ""

	// Mock expectations
	mockStatusManager.EXPECT().AddWorktree(repoName, branch, worktreePath, workspacePath).Return(nil)

	// Execute
	err := wtm.(*realWTM).addWorktreeToStatus(repoName, branch, worktreePath, workspacePath)

	// Assert
	assert.NoError(t, err)
}

func TestAddWorktreeToStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	wtm := NewWTM(createStatusTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	worktreePath := "/home/user/.wtm/repos/github.com/lerenn/example/feature-a"
	workspacePath := ""

	// Mock expectations
	mockStatusManager.EXPECT().AddWorktree(repoName, branch, worktreePath, workspacePath).Return(assert.AnError)

	// Execute
	err := wtm.(*realWTM).addWorktreeToStatus(repoName, branch, worktreePath, workspacePath)

	// Assert
	assert.ErrorIs(t, err, ErrAddWorktreeToStatus)
}

func TestRemoveWorktreeFromStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	wtm := NewWTM(createStatusTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Mock expectations
	mockStatusManager.EXPECT().RemoveWorktree(repoName, branch).Return(nil)

	// Execute
	err := wtm.(*realWTM).removeWorktreeFromStatus(repoName, branch)

	// Assert
	assert.NoError(t, err)
}

func TestRemoveWorktreeFromStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	wtm := NewWTM(createStatusTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Mock expectations
	mockStatusManager.EXPECT().RemoveWorktree(repoName, branch).Return(assert.AnError)

	// Execute
	err := wtm.(*realWTM).removeWorktreeFromStatus(repoName, branch)

	// Assert
	assert.ErrorIs(t, err, ErrRemoveWorktreeFromStatus)
}

func TestGetWorktreeStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	wtm := NewWTM(createStatusTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	expectedRepo := &status.Repository{
		Name:      repoName,
		Branch:    branch,
		Path:      "/home/user/.wtm/repos/github.com/lerenn/example/feature-a",
		Workspace: "",
	}

	// Mock expectations
	mockStatusManager.EXPECT().GetWorktree(repoName, branch).Return(expectedRepo, nil)

	// Execute
	repo, err := wtm.(*realWTM).getWorktreeStatus(repoName, branch)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRepo, repo)
}

func TestGetWorktreeStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	wtm := NewWTM(createStatusTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Mock expectations
	mockStatusManager.EXPECT().GetWorktree(repoName, branch).Return(nil, assert.AnError)

	// Execute
	repo, err := wtm.(*realWTM).getWorktreeStatus(repoName, branch)

	// Assert
	assert.Nil(t, repo)
	assert.ErrorIs(t, err, ErrGetWorktreeStatus)
}

func TestListAllWorktrees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	wtm := NewWTM(createStatusTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Expected repositories
	expectedRepos := []status.Repository{
		{
			Name:      "github.com/lerenn/example",
			Branch:    "feature-a",
			Path:      "/home/user/.wtm/repos/github.com/lerenn/example/feature-a",
			Workspace: "",
		},
		{
			Name:      "github.com/lerenn/other",
			Branch:    "feature-b",
			Path:      "/home/user/.wtm/repos/github.com/lerenn/other/feature-b",
			Workspace: "/home/user/workspace.code-workspace",
		},
	}

	// Mock expectations
	mockStatusManager.EXPECT().ListAllWorktrees().Return(expectedRepos, nil)

	// Execute
	repos, err := wtm.(*realWTM).listAllWorktrees()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestListAllWorktrees_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	wtm := NewWTM(createStatusTestConfig())
	wtm.SetVerbose(true)

	// Override adapters with mocks
	c := wtm.(*realWTM)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Mock expectations
	mockStatusManager.EXPECT().ListAllWorktrees().Return(nil, assert.AnError)

	// Execute
	repos, err := wtm.(*realWTM).listAllWorktrees()

	// Assert
	assert.Nil(t, repos)
	assert.ErrorIs(t, err, ErrListWorktrees)
}
