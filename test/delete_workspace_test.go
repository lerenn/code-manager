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

func TestDeleteWorktree_WorkspaceMode(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "wtm-workspace-delete-test-*")
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

	// Create worktrees first
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

	// Now delete the worktrees
	err = wtmInstance.DeleteWorkTree(branchName, false)
	require.NoError(t, err)

	// Verify worktrees were deleted
	worktrees, _, err = wtmInstance.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 0)

	// Verify worktree directories were removed
	assert.NoDirExists(t, frontendWorktreePath)
	assert.NoDirExists(t, backendWorktreePath)

	// Verify worktree-specific workspace file was removed
	assert.NoFileExists(t, workspaceWorktreePath)

	// Verify status file entries were removed
	statusManager := status.NewManager(fs.NewFS(), cfg)
	allWorktrees, err := statusManager.ListAllWorktrees()
	require.NoError(t, err)
	assert.Len(t, allWorktrees, 0)
}

func TestDeleteWorktree_WorkspaceMode_Force(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "wtm-workspace-delete-force-test-*")
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

	// Create WTM instance
	cfg := &config.Config{
		BasePath:   tempDir,
		StatusFile: filepath.Join(tempDir, "status.yaml"),
	}
	wtmInstance := wtm.NewWTM(cfg)

	// Create worktrees first
	branchName := "feature/test-branch"
	err = wtmInstance.CreateWorkTree(branchName, nil)
	require.NoError(t, err)

	// Verify worktree was created
	worktrees, _, err := wtmInstance.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 1)

	// Verify worktree directory exists
	frontendWorktreePath := filepath.Join(tempDir, "frontend", branchName)
	assert.DirExists(t, frontendWorktreePath)

	// Now delete the worktrees with force
	err = wtmInstance.DeleteWorkTree(branchName, true)
	require.NoError(t, err)

	// Verify worktree was deleted
	worktrees, _, err = wtmInstance.ListWorktrees()
	require.NoError(t, err)
	assert.Len(t, worktrees, 0)

	// Verify worktree directory was removed
	assert.NoDirExists(t, frontendWorktreePath)

	// Verify status file entries were removed
	statusManager := status.NewManager(fs.NewFS(), cfg)
	allWorktrees, err := statusManager.ListAllWorktrees()
	require.NoError(t, err)
	assert.Len(t, allWorktrees, 0)
}
