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
				URL:       repoName,
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
	err := manager.AddWorktree(AddWorktreeParams{
		RepoURL:       repoName,
		Branch:        branch,
		WorktreePath:  worktreePath,
		WorkspacePath: workspacePath,
	})

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
				URL:       repoName,
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
	err := manager.AddWorktree(AddWorktreeParams{
		RepoURL:       repoName,
		Branch:        branch,
		WorktreePath:  worktreePath,
		WorkspacePath: workspacePath,
	})

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
				URL:       repoName,
				Branch:    branch,
				Path:      "/home/user/.wtmrepos/github.com/lerenn/example/feature-a",
				Workspace: "",
			},
			{
				URL:       "github.com/lerenn/other",
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
				URL:       "github.com/lerenn/other",
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
				URL:       "github.com/lerenn/other",
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
		URL:       repoName,
		Branch:    branch,
		Path:      "/home/user/.wtmrepos/github.com/lerenn/example/feature-a",
		Workspace: "",
	}

	// Existing status
	existingStatus := &Status{
		Repositories: []Repository{
			expectedRepo,
			{
				URL:       "github.com/lerenn/other",
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
				URL:       "github.com/lerenn/other",
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
			URL:       "github.com/lerenn/example",
			Branch:    "feature-a",
			Path:      "/home/user/.wtmrepos/github.com/lerenn/example/feature-a",
			Workspace: "",
		},
		{
			URL:       "github.com/lerenn/other",
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

	// Mock expectations for initialization
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(false, nil)
	mockFS.EXPECT().FileLock("/home/user/.wtmstatus.yaml").Return(func() {}, nil)
	mockFS.EXPECT().WriteFileAtomic("/home/user/.wtmstatus.yaml", gomock.Any(), gomock.Any()).Return(nil)

	manager := NewManager(mockFS, cfg)

	assert.NotNil(t, manager)
	assert.Implements(t, (*Manager)(nil), manager)
}

func TestGetWorkspaceWorktrees(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	// Test data
	workspacePath := "/home/user/workspace.code-workspace"
	branchName := "feature-a"

	// Expected repositories for the workspace and branch
	expectedRepos := []Repository{
		{
			URL:       "github.com/lerenn/frontend",
			Branch:    branchName,
			Path:      "/home/user/repos/frontend",
			Workspace: workspacePath,
		},
		{
			URL:       "github.com/lerenn/backend",
			Branch:    branchName,
			Path:      "/home/user/repos/backend",
			Workspace: workspacePath,
		},
	}

	// Existing status with workspace repositories
	existingStatus := &Status{
		Repositories: []Repository{
			// Repositories for the target workspace and branch
			expectedRepos[0],
			expectedRepos[1],
			// Repository for different workspace
			{
				URL:       "github.com/lerenn/other",
				Branch:    "feature-b",
				Path:      "/home/user/repos/other",
				Workspace: "/home/user/other-workspace.code-workspace",
			},
			// Repository for same workspace but different branch
			{
				URL:       "github.com/lerenn/frontend",
				Branch:    "feature-b",
				Path:      "/home/user/repos/frontend-b",
				Workspace: workspacePath,
			},
			// Non-workspace repository
			{
				URL:       "github.com/lerenn/standalone",
				Branch:    "main",
				Path:      "/home/user/repos/standalone",
				Workspace: "",
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations for initialization
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	manager := NewManager(mockFS, cfg).(*realManager)

	// Execute
	repos, err := manager.GetWorkspaceWorktrees(workspacePath, branchName)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedRepos, repos)
}

func TestGetWorkspaceWorktrees_EmptyWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	// Test data
	workspacePath := "/home/user/empty-workspace.code-workspace"
	branchName := "feature-a"

	// Existing status without the target workspace
	existingStatus := &Status{
		Repositories: []Repository{
			{
				URL:       "github.com/lerenn/other",
				Branch:    "feature-b",
				Path:      "/home/user/repos/other",
				Workspace: "/home/user/other-workspace.code-workspace",
			},
			// Non-workspace repository
			{
				URL:       "github.com/lerenn/standalone",
				Branch:    "main",
				Path:      "/home/user/repos/standalone",
				Workspace: "",
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations for initialization
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	manager := NewManager(mockFS, cfg).(*realManager)

	// Execute
	repos, err := manager.GetWorkspaceWorktrees(workspacePath, branchName)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, repos)
}

func TestGetWorkspaceWorktrees_EmptyBranch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	// Test data
	workspacePath := "/home/user/workspace.code-workspace"
	branchName := "non-existent-branch"

	// Existing status with workspace but different branch
	existingStatus := &Status{
		Repositories: []Repository{
			{
				URL:       "github.com/lerenn/frontend",
				Branch:    "feature-a",
				Path:      "/home/user/repos/frontend",
				Workspace: workspacePath,
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations for initialization
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	manager := NewManager(mockFS, cfg).(*realManager)

	// Execute
	repos, err := manager.GetWorkspaceWorktrees(workspacePath, branchName)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, repos)
}

func TestGetWorkspaceBranches(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	// Test data
	workspacePath := "/home/user/workspace.code-workspace"

	// Expected branches for the workspace
	expectedBranches := []string{"feature-a", "feature-b"}

	// Existing status with workspace repositories
	existingStatus := &Status{
		Repositories: []Repository{
			// Repositories for the target workspace
			{
				URL:       "github.com/lerenn/frontend",
				Branch:    "feature-a",
				Path:      "/home/user/repos/frontend",
				Workspace: workspacePath,
			},
			{
				URL:       "github.com/lerenn/backend",
				Branch:    "feature-a",
				Path:      "/home/user/repos/backend",
				Workspace: workspacePath,
			},
			{
				URL:       "github.com/lerenn/frontend",
				Branch:    "feature-b",
				Path:      "/home/user/repos/frontend-b",
				Workspace: workspacePath,
			},
			// Repository for different workspace
			{
				URL:       "github.com/lerenn/other",
				Branch:    "feature-c",
				Path:      "/home/user/repos/other",
				Workspace: "/home/user/other-workspace.code-workspace",
			},
			// Non-workspace repository
			{
				URL:       "github.com/lerenn/standalone",
				Branch:    "main",
				Path:      "/home/user/repos/standalone",
				Workspace: "",
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations for initialization
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	manager := NewManager(mockFS, cfg).(*realManager)

	// Execute
	branches, err := manager.GetWorkspaceBranches(workspacePath)

	// Assert
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedBranches, branches)
}

func TestGetWorkspaceBranches_EmptyWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	// Test data
	workspacePath := "/home/user/empty-workspace.code-workspace"

	// Existing status without the target workspace
	existingStatus := &Status{
		Repositories: []Repository{
			{
				URL:       "github.com/lerenn/other",
				Branch:    "feature-b",
				Path:      "/home/user/repos/other",
				Workspace: "/home/user/other-workspace.code-workspace",
			},
			// Non-workspace repository
			{
				URL:       "github.com/lerenn/standalone",
				Branch:    "main",
				Path:      "/home/user/repos/standalone",
				Workspace: "",
			},
		},
	}

	existingData, _ := yaml.Marshal(existingStatus)

	// Mock expectations for initialization
	mockFS.EXPECT().Exists("/home/user/.wtmstatus.yaml").Return(true, nil)
	mockFS.EXPECT().ReadFile("/home/user/.wtmstatus.yaml").Return(existingData, nil)

	manager := NewManager(mockFS, cfg).(*realManager)

	// Execute
	branches, err := manager.GetWorkspaceBranches(workspacePath)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, branches)
}

func TestComputeWorkspacesMap(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:         mockFS,
		config:     cfg,
		workspaces: make(map[string]map[string][]Repository),
	}

	// Test data
	repositories := []Repository{
		// Workspace1 repositories
		{
			URL:       "github.com/lerenn/frontend",
			Branch:    "feature-a",
			Path:      "/home/user/repos/frontend",
			Workspace: "/home/user/workspace1.code-workspace",
		},
		{
			URL:       "github.com/lerenn/backend",
			Branch:    "feature-a",
			Path:      "/home/user/repos/backend",
			Workspace: "/home/user/workspace1.code-workspace",
		},
		{
			URL:       "github.com/lerenn/frontend",
			Branch:    "feature-b",
			Path:      "/home/user/repos/frontend-b",
			Workspace: "/home/user/workspace1.code-workspace",
		},
		// Workspace2 repositories
		{
			URL:       "github.com/lerenn/other",
			Branch:    "feature-c",
			Path:      "/home/user/repos/other",
			Workspace: "/home/user/workspace2.code-workspace",
		},
		// Non-workspace repository
		{
			URL:       "github.com/lerenn/standalone",
			Branch:    "main",
			Path:      "/home/user/repos/standalone",
			Workspace: "",
		},
	}

	// Execute
	manager.computeWorkspacesMap(repositories)

	// Verify the map structure
	assert.NotNil(t, manager.workspaces)

	// Check workspace1
	workspace1 := manager.workspaces["workspace1"]
	assert.NotNil(t, workspace1)
	assert.Len(t, workspace1["feature-a"], 2)
	assert.Len(t, workspace1["feature-b"], 1)

	// Check workspace2
	workspace2 := manager.workspaces["workspace2"]
	assert.NotNil(t, workspace2)
	assert.Len(t, workspace2["feature-c"], 1)

	// Verify non-workspace repository is not included
	assert.NotContains(t, manager.workspaces, "")
}

func TestGetWorkspaceNameFromPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)

	cfg := &config.Config{
		BasePath:   "/home/user/.wtm",
		StatusFile: "/home/user/.wtmstatus.yaml",
	}

	manager := &realManager{
		fs:         mockFS,
		config:     cfg,
		workspaces: make(map[string]map[string][]Repository),
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
