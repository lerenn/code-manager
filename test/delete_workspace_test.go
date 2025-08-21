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
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDeleteWorktree_WorkspaceMode(t *testing.T) {
	// Create temporary test directory for CM base path
	tempDir, err := os.MkdirTemp("", "cm-workspace-delete-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create separate directory for workspace structure
	workspaceBaseDir, err := os.MkdirTemp("", "cm-workspace-structure-*")
	require.NoError(t, err)
	defer os.RemoveAll(workspaceBaseDir)

	// Create temporary config
	testConfig := &config.Config{
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
	cfg := &config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	cmInstance := cm.NewCM(cfg)

	// Create worktrees first
	branchName := "feature/test-branch"
	err = cmInstance.CreateWorkTree(branchName)
	require.NoError(t, err)

	// Verify worktrees were created
	worktrees, _, err := cmInstance.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 2)

	// Verify worktree directories exist (using the correct path structure)
	helloWorldWorktreePath := filepath.Join(tempDir, "github.com/octocat/Hello-World", "origin", branchName)
	spoonKnifeWorktreePath := filepath.Join(tempDir, "github.com/octocat/Spoon-Knife", "origin", branchName)
	assert.DirExists(t, helloWorldWorktreePath)
	assert.DirExists(t, spoonKnifeWorktreePath)

	// Verify worktree-specific workspace file was created
	workspaceWorktreePath := filepath.Join(tempDir, "workspaces", "test-workspace-feature-test-branch.code-workspace")
	assert.FileExists(t, workspaceWorktreePath)

	// Now delete the worktrees
	err = cmInstance.DeleteWorkTree(branchName, false)
	require.NoError(t, err)

	// Verify worktrees were deleted
	worktrees, _, err = cmInstance.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 0, "Should have no worktrees after deletion")

	// Verify worktree directories were removed
	assert.NoDirExists(t, helloWorldWorktreePath)
	assert.NoDirExists(t, spoonKnifeWorktreePath)

	// Verify worktree-specific workspace file was removed
	assert.NoFileExists(t, workspaceWorktreePath)

	// Verify status file entries were removed
	statusManager := status.NewManager(fs.NewFS(), cfg)
	allWorktrees, err := statusManager.ListAllWorktrees()
	require.NoError(t, err)
	assert.Len(t, allWorktrees, 0, "Should have no worktrees in status file after deletion")
}

func TestDeleteWorktree_WorkspaceMode_Force(t *testing.T) {
	// Create temporary test directory for CM base path
	tempDir, err := os.MkdirTemp("", "cm-workspace-delete-force-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create separate directory for workspace structure
	workspaceBaseDir, err := os.MkdirTemp("", "cm-workspace-structure-*")
	require.NoError(t, err)
	defer os.RemoveAll(workspaceBaseDir)

	// Create temporary config
	testConfig := &config.Config{
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
		},
	}

	workspaceData, err := json.MarshalIndent(workspaceConfig, "", "  ")
	require.NoError(t, err)
	workspacePath := filepath.Join(workspaceDir, "project.code-workspace")
	require.NoError(t, os.WriteFile(workspacePath, workspaceData, 0644))

	// Create repository
	helloWorldDir := filepath.Join(workspaceDir, "Hello-World")
	require.NoError(t, os.MkdirAll(helloWorldDir, 0755))

	// Initialize Git repository
	createTestGitRepo(t, helloWorldDir)

	// Change to workspace directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	require.NoError(t, os.Chdir(workspaceDir))

	// Create CM instance
	cfg := &config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	cmInstance := cm.NewCM(cfg)

	// Create worktrees first
	branchName := "feature/test-branch"
	err = cmInstance.CreateWorkTree(branchName)
	require.NoError(t, err)

	// Verify worktree was created
	worktrees, _, err := cmInstance.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)

	// Verify worktree directory exists (using the correct path structure)
	helloWorldWorktreePath := filepath.Join(tempDir, "github.com/octocat/Hello-World", "origin", branchName)
	assert.DirExists(t, helloWorldWorktreePath)

	// Now delete the worktrees with force
	err = cmInstance.DeleteWorkTree(branchName, true)
	require.NoError(t, err)

	// Verify worktree was deleted
	worktrees, _, err = cmInstance.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 0)

	// Verify worktree directory was removed
	assert.NoDirExists(t, helloWorldWorktreePath)

	// Verify status file entries were removed
	statusManager := status.NewManager(fs.NewFS(), cfg)
	allWorktrees, err := statusManager.ListAllWorktrees()
	require.NoError(t, err)
	assert.Len(t, allWorktrees, 0)
}
