//go:build unit

package cgwt

import (
	"testing"

	"github.com/lerenn/cgwt/pkg/config"
	"github.com/lerenn/cgwt/pkg/fs"
	"github.com/lerenn/cgwt/pkg/git"
	"github.com/lerenn/cgwt/pkg/logger"
	"github.com/lerenn/cgwt/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAddWorktreeToStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cursor/cgwt",
		StatusFile: "/home/user/.cursor/cgwt/status.yaml",
	}

	cgwt := &realCGWT{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatusManager,
		verbose:       true,
		logger:        mockLogger,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	worktreePath := "/home/user/.cursor/cgwt/repos/github.com/lerenn/example/feature-a"
	workspacePath := ""

	// Mock expectations
	mockLogger.EXPECT().Logf("Adding worktree to status: repo=%s, branch=%s, path=%s, workspace=%s", repoName, branch, worktreePath, workspacePath)
	mockStatusManager.EXPECT().AddWorktree(repoName, branch, worktreePath, workspacePath).Return(nil)
	mockLogger.EXPECT().Logf("Successfully added worktree to status")

	// Execute
	err := cgwt.AddWorktreeToStatus(repoName, branch, worktreePath, workspacePath)

	// Assert
	assert.NoError(t, err)
}

func TestAddWorktreeToStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cursor/cgwt",
		StatusFile: "/home/user/.cursor/cgwt/status.yaml",
	}

	cgwt := &realCGWT{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatusManager,
		verbose:       true,
		logger:        mockLogger,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	worktreePath := "/home/user/.cursor/cgwt/repos/github.com/lerenn/example/feature-a"
	workspacePath := ""

	// Mock expectations
	mockLogger.EXPECT().Logf("Adding worktree to status: repo=%s, branch=%s, path=%s, workspace=%s", repoName, branch, worktreePath, workspacePath)
	mockStatusManager.EXPECT().AddWorktree(repoName, branch, worktreePath, workspacePath).Return(assert.AnError)

	// Execute
	err := cgwt.AddWorktreeToStatus(repoName, branch, worktreePath, workspacePath)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add worktree to status")
}

func TestRemoveWorktreeFromStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cursor/cgwt",
		StatusFile: "/home/user/.cursor/cgwt/status.yaml",
	}

	cgwt := &realCGWT{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatusManager,
		verbose:       true,
		logger:        mockLogger,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Mock expectations
	mockLogger.EXPECT().Logf("Removing worktree from status: repo=%s, branch=%s", repoName, branch)
	mockStatusManager.EXPECT().RemoveWorktree(repoName, branch).Return(nil)
	mockLogger.EXPECT().Logf("Successfully removed worktree from status")

	// Execute
	err := cgwt.RemoveWorktreeFromStatus(repoName, branch)

	// Assert
	assert.NoError(t, err)
}

func TestRemoveWorktreeFromStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cursor/cgwt",
		StatusFile: "/home/user/.cursor/cgwt/status.yaml",
	}

	cgwt := &realCGWT{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatusManager,
		verbose:       true,
		logger:        mockLogger,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Mock expectations
	mockLogger.EXPECT().Logf("Removing worktree from status: repo=%s, branch=%s", repoName, branch)
	mockStatusManager.EXPECT().RemoveWorktree(repoName, branch).Return(assert.AnError)

	// Execute
	err := cgwt.RemoveWorktreeFromStatus(repoName, branch)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove worktree from status")
}

func TestGetWorktreeStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cursor/cgwt",
		StatusFile: "/home/user/.cursor/cgwt/status.yaml",
	}

	cgwt := &realCGWT{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatusManager,
		verbose:       true,
		logger:        mockLogger,
	}

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
	mockLogger.EXPECT().Logf("Getting worktree status: repo=%s, branch=%s", repoName, branch)
	mockStatusManager.EXPECT().GetWorktree(repoName, branch).Return(expectedRepo, nil)

	// Execute
	repo, err := cgwt.GetWorktreeStatus(repoName, branch)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRepo, repo)
}

func TestGetWorktreeStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cursor/cgwt",
		StatusFile: "/home/user/.cursor/cgwt/status.yaml",
	}

	cgwt := &realCGWT{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatusManager,
		verbose:       true,
		logger:        mockLogger,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Mock expectations
	mockLogger.EXPECT().Logf("Getting worktree status: repo=%s, branch=%s", repoName, branch)
	mockStatusManager.EXPECT().GetWorktree(repoName, branch).Return(nil, assert.AnError)

	// Execute
	repo, err := cgwt.GetWorktreeStatus(repoName, branch)

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
	mockLogger := logger.NewMockLogger(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cursor/cgwt",
		StatusFile: "/home/user/.cursor/cgwt/status.yaml",
	}

	cgwt := &realCGWT{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatusManager,
		verbose:       true,
		logger:        mockLogger,
	}

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
	mockLogger.EXPECT().Logf("Listing all worktrees")
	mockStatusManager.EXPECT().ListAllWorktrees().Return(expectedRepos, nil)
	mockLogger.EXPECT().Logf("Found %d worktrees", 2)

	// Execute
	repos, err := cgwt.ListAllWorktrees()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestListAllWorktrees_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockLogger := logger.NewMockLogger(ctrl)
	mockStatusManager := status.NewMockManager(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cursor/cgwt",
		StatusFile: "/home/user/.cursor/cgwt/status.yaml",
	}

	cgwt := &realCGWT{
		fs:            mockFS,
		git:           mockGit,
		config:        cfg,
		statusManager: mockStatusManager,
		verbose:       true,
		logger:        mockLogger,
	}

	// Mock expectations
	mockLogger.EXPECT().Logf("Listing all worktrees")
	mockStatusManager.EXPECT().ListAllWorktrees().Return(nil, assert.AnError)

	// Execute
	repos, err := cgwt.ListAllWorktrees()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, repos)
	assert.Contains(t, err.Error(), "failed to list worktrees")
}
