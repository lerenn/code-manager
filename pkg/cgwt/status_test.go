//go:build unit

package cgwt

import (
	"testing"

	"github.com/lerenn/cgwt/pkg/config"
	"github.com/lerenn/cgwt/pkg/fs"
	"github.com/lerenn/cgwt/pkg/git"
	"github.com/lerenn/cgwt/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createStatusTestConfig creates a test configuration for status tests.
func createStatusTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/home/user/.cursor/cgwt",
		StatusFile: "/home/user/.cursor/cgwt/status.yaml",
	}
}

func TestAddWorktreeToStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cgwt := NewCGWT(createStatusTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	worktreePath := "/home/user/.cursor/cgwt/repos/github.com/lerenn/example/feature-a"
	workspacePath := ""

	// Mock expectations
	mockStatusManager.EXPECT().AddWorktree(repoName, branch, worktreePath, workspacePath).Return(nil)

	// Execute
	err := cgwt.(*realCGWT).addWorktreeToStatus(repoName, branch, worktreePath, workspacePath)

	// Assert
	assert.NoError(t, err)
}

func TestAddWorktreeToStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cgwt := NewCGWT(createStatusTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	worktreePath := "/home/user/.cursor/cgwt/repos/github.com/lerenn/example/feature-a"
	workspacePath := ""

	// Mock expectations
	mockStatusManager.EXPECT().AddWorktree(repoName, branch, worktreePath, workspacePath).Return(assert.AnError)

	// Execute
	err := cgwt.(*realCGWT).addWorktreeToStatus(repoName, branch, worktreePath, workspacePath)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add worktree to status")
}

func TestRemoveWorktreeFromStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cgwt := NewCGWT(createStatusTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Mock expectations
	mockStatusManager.EXPECT().RemoveWorktree(repoName, branch).Return(nil)

	// Execute
	err := cgwt.(*realCGWT).removeWorktreeFromStatus(repoName, branch)

	// Assert
	assert.NoError(t, err)
}

func TestRemoveWorktreeFromStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cgwt := NewCGWT(createStatusTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Mock expectations
	mockStatusManager.EXPECT().RemoveWorktree(repoName, branch).Return(assert.AnError)

	// Execute
	err := cgwt.(*realCGWT).removeWorktreeFromStatus(repoName, branch)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove worktree from status")
}

func TestGetWorktreeStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cgwt := NewCGWT(createStatusTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	expectedRepo := &status.Repository{
		Name:      repoName,
		Branch:    branch,
		Path:      "/home/user/.cursor/cgwt/repos/github.com/lerenn/example/feature-a",
		Workspace: "",
	}

	// Mock expectations
	mockStatusManager.EXPECT().GetWorktree(repoName, branch).Return(expectedRepo, nil)

	// Execute
	repo, err := cgwt.(*realCGWT).getWorktreeStatus(repoName, branch)

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

	cgwt := NewCGWT(createStatusTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Mock expectations
	mockStatusManager.EXPECT().GetWorktree(repoName, branch).Return(nil, assert.AnError)

	// Execute
	repo, err := cgwt.(*realCGWT).getWorktreeStatus(repoName, branch)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, repo)
	assert.Contains(t, err.Error(), "failed to get worktree status")
}

func TestListAllWorktrees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cgwt := NewCGWT(createStatusTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Expected repositories
	expectedRepos := []status.Repository{
		{
			Name:      "github.com/lerenn/example",
			Branch:    "feature-a",
			Path:      "/home/user/.cursor/cgwt/repos/github.com/lerenn/example/feature-a",
			Workspace: "",
		},
		{
			Name:      "github.com/lerenn/other",
			Branch:    "feature-b",
			Path:      "/home/user/.cursor/cgwt/repos/github.com/lerenn/other/feature-b",
			Workspace: "/home/user/workspace.code-workspace",
		},
	}

	// Mock expectations
	mockStatusManager.EXPECT().ListAllWorktrees().Return(expectedRepos, nil)

	// Execute
	repos, err := cgwt.(*realCGWT).listAllWorktrees()

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

	cgwt := NewCGWT(createStatusTestConfig())
	cgwt.SetVerbose(true)

	// Override adapters with mocks
	c := cgwt.(*realCGWT)
	c.fs = mockFS
	c.git = mockGit
	c.statusManager = mockStatusManager

	// Mock expectations
	mockStatusManager.EXPECT().ListAllWorktrees().Return(nil, assert.AnError)

	// Execute
	repos, err := cgwt.(*realCGWT).listAllWorktrees()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, repos)
	assert.Contains(t, err.Error(), "failed to list worktrees")
}
