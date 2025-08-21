//go:build unit

package status

import (
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
)

func TestAddWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"
	remote := "origin"
	worktreePath := "/home/user/.cmrepos/github.com/lerenn/example/feature-a"
	workspacePath := ""

	// Expected status file content
	expectedStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-a": {
						Remote: remote,
						Branch: branch,
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return([]byte(`initialized: true
repositories:
  github.com/lerenn/example:
    path: /home/user/.cmrepos/github.com/lerenn/example/origin/main
    remotes:
      origin:
        default_branch: main
    worktrees: {}
workspaces: {}`), nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.AddWorktree(AddWorktreeParams{
		RepoURL:       repoURL,
		Branch:        branch,
		WorktreePath:  worktreePath,
		WorkspacePath: workspacePath,
		Remote:        remote,
	})

	// Assert
	assert.NoError(t, err)
}

func TestAddWorktree_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"
	remote := "origin"

	// Existing status with duplicate
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-a": {
						Remote: remote,
						Branch: branch,
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.AddWorktree(AddWorktreeParams{
		RepoURL: repoURL,
		Branch:  branch,
		Remote:  remote,
	})

	// Assert
	assert.ErrorIs(t, err, ErrWorktreeAlreadyExists)
}

func TestAddWorktree_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"
	remote := "origin"
	worktreePath := "/home/user/.cmrepos/github.com/lerenn/example/origin/feature-a"

	// Existing status without the repository
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations for status file operations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.AddWorktree(AddWorktreeParams{
		RepoURL:      repoURL,
		Branch:       branch,
		Remote:       remote,
		WorktreePath: worktreePath,
	})

	// Assert
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}

func TestRemoveWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"

	// Existing status with the worktree to remove
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-a": {
						Remote: "origin",
						Branch: branch,
					},
					"origin:feature-b": {
						Remote: "origin",
						Branch: "feature-b",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	// Expected status after removal
	expectedStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-b": {
						Remote: "origin",
						Branch: "feature-b",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)
	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.RemoveWorktree(repoURL, branch)

	// Assert
	assert.NoError(t, err)
}

func TestRemoveWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"

	// Existing status without the worktree to remove
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-b": {
						Remote: "origin",
						Branch: "feature-b",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.RemoveWorktree(repoURL, branch)

	// Assert
	assert.ErrorIs(t, err, ErrWorktreeNotFound)
}

func TestRemoveWorktree_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"

	// Existing status without the repository
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.RemoveWorktree(repoURL, branch)

	// Assert
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}

func TestGetWorktree(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"
	expectedWorktree := WorktreeInfo{
		Remote: "origin",
		Branch: branch,
	}

	// Existing status
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-a": expectedWorktree,
					"origin:feature-b": {
						Remote: "origin",
						Branch: "feature-b",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	worktree, err := manager.GetWorktree(repoURL, branch)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, &expectedWorktree, worktree)
}

func TestGetWorktree_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"

	// Existing status without the requested worktree
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-b": {
						Remote: "origin",
						Branch: "feature-b",
					},
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	worktree, err := manager.GetWorktree(repoURL, branch)

	// Assert
	assert.Nil(t, worktree)
	assert.ErrorIs(t, err, ErrWorktreeNotFound)
}

func TestGetWorktree_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	branch := "feature-a"

	// Existing status without the repository
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	worktree, err := manager.GetWorktree(repoURL, branch)

	// Assert
	assert.Nil(t, worktree)
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}

func TestListAllWorktrees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Expected worktrees
	expectedWorktrees := []WorktreeInfo{
		{
			Remote: "origin",
			Branch: "feature-a",
		},
		{
			Remote: "origin",
			Branch: "feature-b",
		},
	}

	// Existing status
	existingStatus := &Status{
		Repositories: map[string]Repository{
			"github.com/lerenn/example": {
				Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
				Remotes: map[string]Remote{
					"origin": {
						DefaultBranch: "main",
					},
				},
				Worktrees: map[string]WorktreeInfo{
					"origin:feature-a": expectedWorktrees[0],
					"origin:feature-b": expectedWorktrees[1],
				},
			},
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	worktrees, err := manager.ListAllWorktrees()

	// Assert
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedWorktrees, worktrees)
}

func TestListAllWorktrees_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Empty status
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	worktrees, err := manager.ListAllWorktrees()

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, worktrees)
}

func TestAddRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	params := AddRepositoryParams{
		Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
	}

	// Expected repository
	expectedRepo := Repository{
		Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
		Worktrees: make(map[string]WorktreeInfo),
	}

	// Existing status
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Expected status after addition
	expectedStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: expectedRepo,
		},
		Workspaces: make(map[string]Workspace),
	}

	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.AddRepository(repoURL, params)

	// Assert
	assert.NoError(t, err)
}

func TestAddRepository_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	params := AddRepositoryParams{
		Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
	}

	// Existing repository
	existingRepo := Repository{
		Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
		Worktrees: make(map[string]WorktreeInfo),
	}

	// Existing status with repository
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: existingRepo,
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.AddRepository(repoURL, params)

	// Assert
	assert.ErrorIs(t, err, ErrRepositoryAlreadyExists)
}

func TestGetRepository(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"
	expectedRepo := Repository{
		Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
		Remotes: map[string]Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
		Worktrees: make(map[string]WorktreeInfo),
	}

	// Existing status
	existingStatus := &Status{
		Repositories: map[string]Repository{
			repoURL: expectedRepo,
		},
		Workspaces: make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	repo, err := manager.GetRepository(repoURL)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, &expectedRepo, repo)
}

func TestGetRepository_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	repoURL := "github.com/lerenn/example"

	// Existing status without the repository
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	repo, err := manager.GetRepository(repoURL)

	// Assert
	assert.Nil(t, repo)
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}

func TestListRepositories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Expected repositories
	expectedRepos := map[string]Repository{
		"github.com/lerenn/example": {
			Path: "/home/user/.cmrepos/github.com/lerenn/example/origin/main",
			Remotes: map[string]Remote{
				"origin": {
					DefaultBranch: "main",
				},
			},
			Worktrees: make(map[string]WorktreeInfo),
		},
		"github.com/lerenn/other": {
			Path: "/home/user/repos/other",
			Remotes: map[string]Remote{
				"origin": {
					DefaultBranch: "master",
				},
			},
			Worktrees: make(map[string]WorktreeInfo),
		},
	}

	// Existing status
	existingStatus := &Status{
		Repositories: expectedRepos,
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	repos, err := manager.ListRepositories()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestAddWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	workspacePath := "/home/user/workspace.code-workspace"
	params := AddWorkspaceParams{
		Worktree:     "origin:feature-a",
		Repositories: []string{"github.com/lerenn/example", "github.com/lerenn/other"},
	}

	// Expected status file content
	expectedStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspacePath: {
				Worktree:     params.Worktree,
				Repositories: params.Repositories,
			},
		},
	}

	expectedData, _ := yaml.Marshal(expectedStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return([]byte(`initialized: true
repositories: {}
workspaces: {}`), nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", expectedData, gomock.Any()).Return(nil)

	// Execute
	err := manager.AddWorkspace(workspacePath, params)

	// Assert
	assert.NoError(t, err)
}

func TestAddWorkspace_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	workspacePath := "/home/user/workspace.code-workspace"
	params := AddWorkspaceParams{
		Worktree:     "origin:feature-a",
		Repositories: []string{"github.com/lerenn/example"},
	}

	// Existing status with duplicate workspace
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspacePath: {
				Worktree:     "origin:feature-b",
				Repositories: []string{"github.com/lerenn/other"},
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	err := manager.AddWorkspace(workspacePath, params)

	// Assert
	assert.ErrorIs(t, err, ErrWorkspaceAlreadyExists)
}

func TestGetWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	workspacePath := "/home/user/workspace.code-workspace"
	expectedWorkspace := Workspace{
		Worktree:     "origin:feature-a",
		Repositories: []string{"github.com/lerenn/example", "github.com/lerenn/other"},
	}

	// Existing status
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces: map[string]Workspace{
			workspacePath: expectedWorkspace,
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	workspace, err := manager.GetWorkspace(workspacePath)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, &expectedWorkspace, workspace)
}

func TestGetWorkspace_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Test data
	workspacePath := "/home/user/workspace.code-workspace"

	// Existing status without the workspace
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   make(map[string]Workspace),
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	workspace, err := manager.GetWorkspace(workspacePath)

	// Assert
	assert.Nil(t, workspace)
	assert.ErrorIs(t, err, ErrWorkspaceNotFound)
}

func TestListWorkspaces(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:     mockFS,
		config: cfg,
	}

	// Expected workspaces
	expectedWorkspaces := map[string]Workspace{
		"/home/user/workspace1.code-workspace": {
			Worktree:     "origin:feature-a",
			Repositories: []string{"github.com/lerenn/example"},
		},
		"/home/user/workspace2.code-workspace": {
			Worktree:     "origin:feature-b",
			Repositories: []string{"github.com/lerenn/other"},
		},
	}

	// Existing status
	existingStatus := &Status{
		Repositories: make(map[string]Repository),
		Workspaces:   expectedWorkspaces,
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.cmstatus.yaml").Return(existingData, nil)

	// Execute
	workspaces, err := manager.ListWorkspaces()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedWorkspaces, workspaces)
}

func TestNewManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	// Mock expectations for initialization
	mockFS.EXPECT().Exists("/home/user/.cmstatus.yaml").Return(false, nil)
	mockFS.EXPECT().FileLock("/home/user/.cmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.cmstatus.yaml", gomock.Any(), gomock.Any()).Return(nil)

	manager := NewManager(mockFS, cfg)

	assert.NotNil(t, manager)
	assert.Implements(t, (*Manager)(nil), manager)
}

func TestGetWorkspaceNameFromPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.cm",
		StatusFile: "/home/user/.cmstatus.yaml",
	}

	manager := &realManager{
		fs:         mockFS,
		config:     cfg,
		workspaces: make(map[string]map[string][]WorktreeInfo),
	}

	// Test cases
	testCases := []struct {
		path     string
		expected string
	}{
		{
			path:     "/home/user/workspace.code-workspace",
			expected: "workspace",
		},
		{
			path:     "/path/to/my-project.code-workspace",
			expected: "my-project",
		},
		{
			path:     "simple.code-workspace",
			expected: "simple",
		},
		{
			path:     "/home/user/no-extension",
			expected: "no-extension",
		},
		{
			path:     "",
			expected: ".",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			result := manager.getWorkspaceNameFromPath(tc.path)
			assert.Equal(t, tc.expected, result)
		})
	}
}
