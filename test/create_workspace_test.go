//go:build e2e

package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lerenn/wtm/pkg/config"
	"github.com/lerenn/wtm/pkg/fs"
	"github.com/lerenn/wtm/pkg/status"
	"github.com/lerenn/wtm/pkg/wtm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestCreateWorktree_WorkspaceMode(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "wtm-workspace-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

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
	workspaceDir := filepath.Join(tempDir, "workspace")
	require.NoError(t, os.MkdirAll(workspaceDir, 0755))

	// Create workspace file
	workspaceConfig := &wtm.WorkspaceConfig{
		Folders: []wtm.WorkspaceFolder{
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

	// Create WTM instance
	cfg := &config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	wtmInstance := wtm.NewWTM(cfg)
	wtmInstance.SetVerbose(true)

	// Create worktrees
	branchName := "feature/test-branch"
	err = wtmInstance.CreateWorkTree(branchName, nil)
	require.NoError(t, err)

	// Verify worktrees were created
	worktrees, _, err := wtmInstance.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 2)

	// Verify worktree directories exist
	frontendWorktreePath := filepath.Join(tempDir, "frontend", branchName)
	backendWorktreePath := filepath.Join(tempDir, "backend", branchName)
	assert.DirExists(t, frontendWorktreePath)
	assert.DirExists(t, backendWorktreePath)

	// Verify worktree-specific workspace file was created
	workspaceWorktreePath := filepath.Join(tempDir, "workspaces", "project-feature-test-branch.code-workspace")
	assert.FileExists(t, workspaceWorktreePath)

	// Verify worktree-specific workspace file content
	workspaceWorktreeData, err := os.ReadFile(workspaceWorktreePath)
	require.NoError(t, err)

	var worktreeWorkspaceConfig wtm.WorkspaceConfig
	require.NoError(t, json.Unmarshal(workspaceWorktreeData, &worktreeWorkspaceConfig))

	assert.Equal(t, "project-feature-test-branch", worktreeWorkspaceConfig.Name)
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
	assert.Equal(t, workspacePath, frontendWorktree.Workspace)
	assert.Equal(t, workspacePath, backendWorktree.Workspace)
}
