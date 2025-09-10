//go:build e2e

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCreateWorktree_WorkspaceMode(t *testing.T) {
	// Create temporary test directory for CM base path
	tempDir, err := os.MkdirTemp("", "cm-workspace-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create separate directory for workspace structure
	workspaceBaseDir, err := os.MkdirTemp("", "cm-workspace-structure-*")
	require.NoError(t, err)
	defer os.RemoveAll(workspaceBaseDir)

	// Create temporary config
	testConfig := config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	configPath := filepath.Join(tempDir, "config.yaml")
	configData, err := yaml.Marshal(testConfig)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	// Create workspace structure
	workspaceDir := filepath.Join(workspaceBaseDir, "workspace")
	require.NoError(t, os.MkdirAll(workspaceDir, 0755))

	// Create workspace file
	workspaceConfig := &workspace.Config{
		Name: "test-workspace",
		Folders: []workspace.Folder{
			{Name: "Hello-World", Path: "./Hello-World"},
			{Name: "Spoon-Knife", Path: "./Spoon-Knife"},
		},
	}

	workspaceData, err := json.MarshalIndent(workspaceConfig, "", "  ")
	require.NoError(t, err)
	workspacePath := filepath.Join(workspaceDir, "project.code-workspace")
	require.NoError(t, os.WriteFile(workspacePath, workspaceData, 0644))

	// Create repositories
	helloWorldDir := filepath.Join(workspaceDir, "Hello-World")
	spoonKnifeDir := filepath.Join(workspaceDir, "Spoon-Knife")
	require.NoError(t, os.MkdirAll(helloWorldDir, 0755))
	require.NoError(t, os.MkdirAll(spoonKnifeDir, 0755))

	// Initialize Git repositories
	createTestGitRepo(t, helloWorldDir)
	createTestGitRepo(t, spoonKnifeDir)

	// Change to workspace directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	require.NoError(t, os.Chdir(workspaceDir))

	// Create CM instance
	cfg := config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: cfg,
	})

	require.NoError(t, err)
	// Create worktrees
	branchName := "feature/test-branch"
	err = cmInstance.CreateWorkTree(branchName)
	require.NoError(t, err)

	// Verify worktrees were created
	worktrees, _, err := cmInstance.ListWorktrees(false)
	require.NoError(t, err)
	assert.Len(t, worktrees, 2)

	// Verify worktree directories exist (new structure: repo/origin/branch)
	helloWorldWorktreePath := filepath.Join(tempDir, "github.com/octocat/Hello-World", "origin", branchName)
	spoonKnifeWorktreePath := filepath.Join(tempDir, "github.com/octocat/Spoon-Knife", "origin", branchName)
	assert.DirExists(t, helloWorldWorktreePath)
	assert.DirExists(t, spoonKnifeWorktreePath)

	// Verify worktree-specific workspace file was created
	workspaceWorktreePath := filepath.Join(tempDir, "workspaces", "test-workspace-feature-test-branch.code-workspace")
	assert.FileExists(t, workspaceWorktreePath)

	// Verify worktree-specific workspace file content
	workspaceWorktreeData, err := os.ReadFile(workspaceWorktreePath)
	require.NoError(t, err)

	var worktreeWorkspaceConfig workspace.Config
	require.NoError(t, json.Unmarshal(workspaceWorktreeData, &worktreeWorkspaceConfig))

	assert.Equal(t, "test-workspace-feature-test-branch", worktreeWorkspaceConfig.Name)
	assert.Len(t, worktreeWorkspaceConfig.Folders, 2)
	assert.Equal(t, "Hello-World", worktreeWorkspaceConfig.Folders[0].Name)
	assert.Equal(t, helloWorldWorktreePath, worktreeWorkspaceConfig.Folders[0].Path)
	assert.Equal(t, "Spoon-Knife", worktreeWorkspaceConfig.Folders[1].Name)
	assert.Equal(t, spoonKnifeWorktreePath, worktreeWorkspaceConfig.Folders[1].Path)

	// Verify status file entries
	statusManager := status.NewManager(fs.NewFS(), cfg)
	allWorktrees, err := statusManager.ListAllWorktrees()
	require.NoError(t, err)
	assert.Len(t, allWorktrees, 2)

	// Find Hello-World and Spoon-Knife worktrees
	var helloWorldWorktree, spoonKnifeWorktree *status.WorktreeInfo
	foundHelloWorld := false
	foundSpoonKnife := false

	for _, worktree := range allWorktrees {
		if worktree.Branch == branchName {
			if !foundHelloWorld {
				helloWorldWorktree = &worktree
				foundHelloWorld = true
			} else if !foundSpoonKnife {
				spoonKnifeWorktree = &worktree
				foundSpoonKnife = true
			}
		}
	}

	require.NotNil(t, helloWorldWorktree, "Should have Hello-World worktree")
	require.NotNil(t, spoonKnifeWorktree, "Should have Spoon-Knife worktree")
	assert.Equal(t, branchName, helloWorldWorktree.Branch)
	assert.Equal(t, branchName, spoonKnifeWorktree.Branch)
	assert.Equal(t, "origin", helloWorldWorktree.Remote, "Should have origin remote")
	assert.Equal(t, "origin", spoonKnifeWorktree.Remote, "Should have origin remote")
}

func TestCreateWorkspace_EndToEnd(t *testing.T) {
	// Create temporary test directory for CM base path
	tempDir, err := os.MkdirTemp("", "cm-workspace-create-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create separate directory for repositories
	reposDir, err := os.MkdirTemp("", "cm-repos-*")
	require.NoError(t, err)
	defer os.RemoveAll(reposDir)

	// Create test repository
	repoDir := filepath.Join(reposDir, "Hello-World")
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	// Initialize Git repository
	createTestGitRepo(t, repoDir)

	// Create CM instance
	cfg := config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: cfg,
	})
	require.NoError(t, err)

	// Test 1: Create workspace with absolute path
	workspaceName := "test-workspace-absolute"
	params := cm.CreateWorkspaceParams{
		WorkspaceName: workspaceName,
		Repositories:  []string{repoDir},
	}

	err = cmInstance.CreateWorkspace(params)
	require.NoError(t, err)

	// Verify workspace was created in status file
	statusManager := status.NewManager(fs.NewFS(), cfg)
	workspace, err := statusManager.GetWorkspace(workspaceName)
	require.NoError(t, err)
	require.NotNil(t, workspace)
	assert.Len(t, workspace.Repositories, 1)
	assert.Contains(t, workspace.Repositories, "https://github.com/octocat/Hello-World.git")

	// Test 2: Verify repositories were added to status file
	repositories, err := statusManager.ListRepositories()
	require.NoError(t, err)
	assert.Len(t, repositories, 1)
	assert.Contains(t, repositories, "https://github.com/octocat/Hello-World.git")
}

func TestCreateWorkspace_EndToEnd_ErrorCases(t *testing.T) {
	// Create temporary test directory for CM base path
	tempDir, err := os.MkdirTemp("", "cm-workspace-create-error-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create CM instance
	cfg := config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: cfg,
	})
	require.NoError(t, err)

	// Test 1: Create workspace with empty name
	params := cm.CreateWorkspaceParams{
		WorkspaceName: "",
		Repositories:  []string{"/some/path"},
	}

	err = cmInstance.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cm.ErrInvalidWorkspaceName)

	// Test 2: Create workspace with no repositories
	params = cm.CreateWorkspaceParams{
		WorkspaceName: "empty-workspace",
		Repositories:  []string{},
	}

	err = cmInstance.CreateWorkspace(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one repository must be specified")

	// Test 3: Create workspace with non-existent repository
	params = cm.CreateWorkspaceParams{
		WorkspaceName: "invalid-workspace",
		Repositories:  []string{"/non/existent/path"},
	}

	err = cmInstance.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cm.ErrRepositoryNotFound)

	// Test 4: Create workspace with duplicate repositories
	// Create a test repository first
	repoDir, err := os.MkdirTemp("", "cm-test-repo-*")
	require.NoError(t, err)
	defer os.RemoveAll(repoDir)

	createTestGitRepo(t, repoDir)

	params = cm.CreateWorkspaceParams{
		WorkspaceName: "duplicate-workspace",
		Repositories:  []string{repoDir, repoDir},
	}

	err = cmInstance.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cm.ErrDuplicateRepository)

	// Test 5: Create workspace that already exists
	// First create a workspace
	params = cm.CreateWorkspaceParams{
		WorkspaceName: "existing-workspace",
		Repositories:  []string{repoDir},
	}

	err = cmInstance.CreateWorkspace(params)
	require.NoError(t, err)

	// Try to create the same workspace again
	err = cmInstance.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cm.ErrWorkspaceAlreadyExists)
}

func TestCreateWorkspace_EndToEnd_InvalidRepository(t *testing.T) {
	// Create temporary test directory for CM base path
	tempDir, err := os.MkdirTemp("", "cm-workspace-create-invalid-repo-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create CM instance
	cfg := config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: cfg,
	})
	require.NoError(t, err)

	// Test 1: Create workspace with directory that exists but is not a Git repository
	nonGitDir, err := os.MkdirTemp("", "cm-non-git-repo-*")
	require.NoError(t, err)
	defer os.RemoveAll(nonGitDir)

	// Create a regular directory (not a Git repository)
	err = os.WriteFile(filepath.Join(nonGitDir, "some-file.txt"), []byte("content"), 0644)
	require.NoError(t, err)

	params := cm.CreateWorkspaceParams{
		WorkspaceName: "invalid-repo-workspace",
		Repositories:  []string{nonGitDir},
	}

	err = cmInstance.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cm.ErrInvalidRepository)

	// Test 2: Create workspace with file instead of directory
	testFile, err := os.CreateTemp("", "cm-test-file-*")
	require.NoError(t, err)
	defer os.Remove(testFile.Name())
	defer os.RemoveAll(testFile.Name())

	params = cm.CreateWorkspaceParams{
		WorkspaceName: "file-workspace",
		Repositories:  []string{testFile.Name()},
	}

	err = cmInstance.CreateWorkspace(params)
	assert.Error(t, err)
	assert.ErrorIs(t, err, cm.ErrInvalidRepository)
}

func TestCreateWorkspace_EndToEnd_RelativePathResolution(t *testing.T) {
	// Create temporary test directory for CM base path
	tempDir, err := os.MkdirTemp("", "cm-workspace-create-relative-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create separate directory for repositories
	reposDir, err := os.MkdirTemp("", "cm-repos-relative-*")
	require.NoError(t, err)
	defer os.RemoveAll(reposDir)

	// Create test repository with specific name to get unique remote URL
	repoDir := filepath.Join(reposDir, "Hello-World")
	require.NoError(t, os.MkdirAll(repoDir, 0755))
	createTestGitRepo(t, repoDir)

	// Create CM instance
	cfg := config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: cfg,
	})
	require.NoError(t, err)

	// Test relative path resolution from different working directories
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	// Change to repos directory
	require.NoError(t, os.Chdir(reposDir))

	// Test 1: Simple relative path
	params := cm.CreateWorkspaceParams{
		WorkspaceName: "relative-workspace-1",
		Repositories:  []string{"./Hello-World"},
	}

	err = cmInstance.CreateWorkspace(params)
	require.NoError(t, err)

	// Verify workspace was created
	statusManager := status.NewManager(fs.NewFS(), cfg)
	workspace, err := statusManager.GetWorkspace("relative-workspace-1")
	require.NoError(t, err)
	require.NotNil(t, workspace)
	assert.Len(t, workspace.Repositories, 1)
	assert.Contains(t, workspace.Repositories, "https://github.com/octocat/Hello-World.git")
}
