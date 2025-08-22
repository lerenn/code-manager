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
	cmInstance.SetVerbose(true)

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
