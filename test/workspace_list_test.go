//go:build e2e

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/mode/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestListWorktrees_WorkspaceMode(t *testing.T) {
	// Create temporary test directory for CM base path
	tempDir, err := os.MkdirTemp("", "cm-workspace-list-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create separate directory for workspace structure
	workspaceBaseDir, err := os.MkdirTemp("", "cm-workspace-structure-*")
	require.NoError(t, err)
	defer os.RemoveAll(workspaceBaseDir)

	// Create temporary config
	testConfig := config.Config{
		RepositoriesDir: tempDir,
		StatusFile:      filepath.Join(tempDir, "status.yaml"),
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
		RepositoriesDir: tempDir,
		StatusFile:      filepath.Join(tempDir, "status.yaml"),
	}
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: cfg,
	})

	require.NoError(t, err)
	// Initially, no worktrees should exist
	worktrees, _, err := cmInstance.ListWorktrees(false)
	require.NoError(t, err)
	assert.Len(t, worktrees, 0)

	// Create worktrees
	branchName1 := "feature/branch1"
	branchName2 := "feature/branch2"

	err = cmInstance.CreateWorkTree(branchName1)
	require.NoError(t, err)

	err = cmInstance.CreateWorkTree(branchName2)
	require.NoError(t, err)

	// Verify worktrees are listed
	worktrees, _, err = cmInstance.ListWorktrees(false)
	require.NoError(t, err)
	assert.Len(t, worktrees, 4) // 2 repositories × 2 branches

	// Verify all worktrees have the correct branches and remotes
	for _, worktree := range worktrees {
		assert.Contains(t, []string{branchName1, branchName2}, worktree.Branch)
		assert.Equal(t, "origin", worktree.Remote, "Should have origin remote")
	}

	// Delete one branch
	err = cmInstance.DeleteWorkTree(branchName1, true)
	require.NoError(t, err)

	// Verify only the remaining worktrees are listed
	worktrees, _, err = cmInstance.ListWorktrees(false)
	require.NoError(t, err)
	assert.Len(t, worktrees, 2) // 2 repositories × 1 branch

	// Verify all remaining worktrees have the correct branch
	for _, worktree := range worktrees {
		assert.Equal(t, branchName2, worktree.Branch)
		assert.Equal(t, "origin", worktree.Remote, "Should have origin remote")
	}
}

func TestListWorktrees_WorkspaceMode_Empty(t *testing.T) {
	// Create temporary test directory for CM base path
	tempDir, err := os.MkdirTemp("", "cm-workspace-list-empty-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create separate directory for workspace structure
	workspaceBaseDir, err := os.MkdirTemp("", "cm-workspace-structure-*")
	require.NoError(t, err)
	defer os.RemoveAll(workspaceBaseDir)

	// Create temporary config
	testConfig := config.Config{
		RepositoriesDir: tempDir,
		StatusFile:      filepath.Join(tempDir, "status.yaml"),
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
			{Name: "Frontend", Path: "./frontend"},
		},
	}

	workspaceData, err := json.MarshalIndent(workspaceConfig, "", "  ")
	require.NoError(t, err)
	workspacePath := filepath.Join(workspaceDir, "project.code-workspace")
	require.NoError(t, os.WriteFile(workspacePath, workspaceData, 0644))

	// Create repository
	frontendDir := filepath.Join(workspaceDir, "frontend")
	require.NoError(t, os.MkdirAll(frontendDir, 0755))

	// Initialize Git repository
	createTestGitRepo(t, frontendDir)

	// Change to workspace directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	require.NoError(t, os.Chdir(workspaceDir))

	// Create CM instance
	cfg := config.Config{
		RepositoriesDir: tempDir,
		StatusFile:      filepath.Join(tempDir, "status.yaml"),
	}
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: cfg,
	})

	require.NoError(t, err)
	// Initially, no worktrees should exist
	worktrees, _, err := cmInstance.ListWorktrees(false)
	require.NoError(t, err)
	assert.Len(t, worktrees, 0)

	// Create a worktree
	branchName := "feature/test-branch"
	err = cmInstance.CreateWorkTree(branchName, cm.CreateWorkTreeOpts{
		WorkspaceName: "test-workspace",
	})
	require.NoError(t, err)

	// Verify worktree is listed
	worktrees, _, err = cmInstance.ListWorktrees(false)
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)
	assert.Equal(t, branchName, worktrees[0].Branch)
	assert.Equal(t, "origin", worktrees[0].Remote, "Should have origin remote")

	// Delete the worktree
	err = cmInstance.DeleteWorkTree(branchName, true)
	require.NoError(t, err)

	// Verify no worktrees are listed
	worktrees, _, err = cmInstance.ListWorktrees(false)
	require.NoError(t, err)
	assert.Len(t, worktrees, 0)
}
