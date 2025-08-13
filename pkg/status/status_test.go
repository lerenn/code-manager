//go:build unit

package status

import (
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
)

func TestAddWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	worktreePath := "/home/user/.wtmrepos/github.com/lerenn/example/feature-a"
	workspacePath := ""

	// Expected status file content
	expectedStatus := &Status{
		Repositories: []Repository{
			{
				Name:      repoName,
				Branch:    branch,
				Path:      worktreePath,
				Workspace: workspacePath,
			},
		},
	}

	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(false, nil)
	mockFS.EXPECT().FileLock("/home/user/.wtmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.wtmstatus.yaml", gomock.Any(), gomock.Any()).Return(nil)
	mockFS.EXPECT().FileLock("/home/user/.wtmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.wtmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.AddWorktree(repoName, branch, worktreePath, workspacePath)

	// Assert
	assert.NoError(t, err)
}

func TestAddWorktree_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	worktreePath := "/home/user/.wtmrepos/github.com/lerenn/example/feature-a"
	workspacePath := ""

	// Existing status with duplicate
	existingStatus := &Status{
		Repositories: []Repository{
			{
				Name:      repoName,
				Branch:    branch,
				Path:      "/existing/path",
				Workspace: "",
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.AddWorktree(repoName, branch, worktreePath, workspacePath)

	// Assert
	assert.ErrorIs(t, err, ErrWorktreeAlreadyExists)
}

func TestRemoveWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Existing status with the worktree to remove
	existingStatus := &Status{
		Repositories: []Repository{
			{
				Name:      repoName,
				Branch:    branch,
				Path:      "/home/user/.wtmrepos/github.com/lerenn/example/feature-a",
				Workspace: "",
			},
			{
				Name:      "github.com/lerenn/other",
				Branch:    "feature-b",
				Path:      "/home/user/.wtmrepos/github.com/lerenn/other/feature-b",
				Workspace: "",
			},
		},
	}

	// Expected status after removal
	expectedStatus := &Status{
		Repositories: []Repository{
			{
				Name:      "github.com/lerenn/other",
				Branch:    "feature-b",
				Path:      "/home/user/.wtmrepos/github.com/lerenn/other/feature-b",
				Workspace: "",
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)
	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)
	mockFS.EXPECT().FileLock("/home/user/.wtmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.wtmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.RemoveWorktree(repoName, branch)

	// Assert
	assert.NoError(t, err)
}

func TestRemoveWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Existing status without the worktree to remove
	existingStatus := &Status{
		Repositories: []Repository{
			{
				Name:      "github.com/lerenn/other",
				Branch:    "feature-b",
				Path:      "/home/user/.wtmrepos/github.com/lerenn/other/feature-b",
				Workspace: "",
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.RemoveWorktree(repoName, branch)

	// Assert
	assert.ErrorIs(t, err, ErrWorktreeNotFound)
}

func TestGetWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"
	expectedRepo := Repository{
		Name:      repoName,
		Branch:    branch,
		Path:      "/home/user/.wtmrepos/github.com/lerenn/example/feature-a",
		Workspace: "",
	}

	// Existing status
	existingStatus := &Status{
		Repositories: []Repository{
			expectedRepo,
			{
				Name:      "github.com/lerenn/other",
				Branch:    "feature-b",
				Path:      "/home/user/.wtmrepos/github.com/lerenn/other/feature-b",
				Workspace: "",
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	// Execute
	repo, err := manager.GetWorktree(repoName, branch)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, &expectedRepo, repo)
}

func TestGetWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoName := "github.com/lerenn/example"
	branch := "feature-a"

	// Existing status without the requested worktree
	existingStatus := &Status{
		Repositories: []Repository{
			{
				Name:      "github.com/lerenn/other",
				Branch:    "feature-b",
				Path:      "/home/user/.wtmrepos/github.com/lerenn/other/feature-b",
				Workspace: "",
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	// Execute
	repo, err := manager.GetWorktree(repoName, branch)

	// Assert
	assert.Nil(t, repo)
	assert.ErrorIs(t, err, ErrWorktreeNotFound)
}

func TestListAllWorktrees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Expected repositories
	expectedRepos := []Repository{
		{
			Name:      "github.com/lerenn/example",
			Branch:    "feature-a",
			Path:      "/home/user/.wtmrepos/github.com/lerenn/example/feature-a",
			Workspace: "",
		},
		{
			Name:      "github.com/lerenn/other",
			Branch:    "feature-b",
			Path:      "/home/user/.wtmrepos/github.com/lerenn/other/feature-b",
			Workspace: "/home/user/workspace.code-workspace",
		},
	}

	// Existing status
	existingStatus := &Status{
		Repositories: expectedRepos,
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	// Execute
	repos, err := manager.ListAllWorktrees()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestListAllWorktrees_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Empty status
	existingStatus := &Status{
		Repositories: []Repository{},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	// Execute
	repos, err := manager.ListAllWorktrees()

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, repos)
}

func TestNewManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := NewManager(mockFS, cfg)

	assert.NotNil(t, manager)
	assert.Implements(t, (*Manager)(nil), manager)
}
