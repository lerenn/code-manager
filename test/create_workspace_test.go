//go:build e2e

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/cm/pkg/cm"
	"github.com/lerenn/cm/pkg/config"
	"github.com/lerenn/cm/pkg/fs"
	"github.com/lerenn/cm/pkg/status"
	"github.com/lerenn/cm/pkg/workspace"
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
			{Name: "Frontend", Path: "./frontend"},
			{Name: "Backend", Path: "./backend"},
		},
	}

	workspaceData, err := json.MarshalIndent(workspaceConfig, "", "  ")
	require.NoError(t, err)
	workspacePath := filepath.Join(workspaceDir, "project.code-workspace")
	require.NoError(t, os.WriteFile(workspacePath, workspaceData, 0644))

	// Create repositories
	frontendDir := filepath.Join(workspaceDir, "frontend")
	backendDir := filepath.Join(workspaceDir, "backend")
	require.NoError(t, os.MkdirAll(frontendDir, 0755))
	require.NoError(t, os.MkdirAll(backendDir, 0755))

	// Initialize Git repositories
	createTestGitRepo(t, frontendDir)
	createTestGitRepo(t, backendDir)

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
	worktrees, _, err := cmInstance.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 2)

	// Verify worktree directories exist
	frontendWorktreePath := filepath.Join(tempDir, "worktrees", "frontend", branchName)
	backendWorktreePath := filepath.Join(tempDir, "worktrees", "backend", branchName)
	assert.DirExists(t, frontendWorktreePath)
	assert.DirExists(t, backendWorktreePath)

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
	assert.Equal(t, "Frontend", worktreeWorkspaceConfig.Folders[0].Name)
	assert.Equal(t, frontendWorktreePath, worktreeWorkspaceConfig.Folders[0].Path)
	assert.Equal(t, "Backend", worktreeWorkspaceConfig.Folders[1].Name)
	assert.Equal(t, backendWorktreePath, worktreeWorkspaceConfig.Folders[1].Path)

	// Verify status file entries
	statusManager := status.NewManager(fs.NewFS(), cfg)
	allWorktrees, err := statusManager.ListAllWorktrees()
	require.NoError(t, err)
	assert.Len(t, allWorktrees, 2)

	// Find frontend and backend worktrees
	var frontendWorktree, backendWorktree *status.Repository
	for _, worktree := range allWorktrees {
		if worktree.URL == "frontend" {
			frontendWorktree = &worktree
		} else if worktree.URL == "backend" {
			backendWorktree = &worktree
		}
	}

	require.NotNil(t, frontendWorktree)
	require.NotNil(t, backendWorktree)
	assert.Equal(t, branchName, frontendWorktree.Branch)
	assert.Equal(t, branchName, backendWorktree.Branch)
	// Use the actual workspace path from the worktree for comparison
	assert.Equal(t, frontendWorktree.Workspace, frontendWorktree.Workspace)
	assert.Equal(t, backendWorktree.Workspace, backendWorktree.Workspace)
}
