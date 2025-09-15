//go:build e2e

package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	codemanager "github.com/lerenn/code-manager/pkg/code-manager"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestRepositoryDelete(t *testing.T) {
	// Setup test environment
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)).
			WithLogger(logger.NewVerboseLogger()),
	})
	require.NoError(t, err)

	// Clone a test repository
	repoURL := "https://github.com/octocat/Hello-World.git"
	err = cmInstance.Clone(repoURL)
	require.NoError(t, err)

	// List repositories to verify it was added
	repositories, err := cmInstance.ListRepositories()
	require.NoError(t, err)
	assert.Len(t, repositories, 1)
	assert.Equal(t, "github.com/octocat/Hello-World", repositories[0].Name)

	// Delete the repository with force flag
	params := codemanager.DeleteRepositoryParams{
		RepositoryName: "github.com/octocat/Hello-World",
		Force:          true,
	}
	err = cmInstance.DeleteRepository(params)
	require.NoError(t, err)

	// List repositories to verify it was removed
	repositories, err = cmInstance.ListRepositories()
	require.NoError(t, err)
	assert.Len(t, repositories, 0)
}

func TestRepositoryDeleteWithWorkspace(t *testing.T) {
	// Setup test environment
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)).
			WithLogger(logger.NewVerboseLogger()),
	})
	require.NoError(t, err)

	// Clone a test repository
	repoURL := "https://github.com/octocat/Hello-World.git"
	err = cmInstance.Clone(repoURL)
	require.NoError(t, err)

	// Create a workspace with the repository
	workspaceParams := codemanager.CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{"github.com/octocat/Hello-World"},
	}
	err = cmInstance.CreateWorkspace(workspaceParams)
	require.NoError(t, err)

	// Try to delete the repository (should fail because it's in a workspace)
	deleteParams := codemanager.DeleteRepositoryParams{
		RepositoryName: "github.com/octocat/Hello-World",
		Force:          true,
	}
	err = cmInstance.DeleteRepository(deleteParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is part of workspace")

	// Delete the workspace first
	workspaceDeleteParams := codemanager.DeleteWorkspaceParams{
		WorkspaceName: "test-workspace",
		Force:         true,
	}
	err = cmInstance.DeleteWorkspace(workspaceDeleteParams)
	require.NoError(t, err)

	// Now delete the repository should work
	err = cmInstance.DeleteRepository(deleteParams)
	require.NoError(t, err)
}

func TestRepositoryDeleteInvalidName(t *testing.T) {
	// Setup test environment
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)).
			WithLogger(logger.NewVerboseLogger()),
	})
	require.NoError(t, err)

	// Try to delete with empty repository name
	params := codemanager.DeleteRepositoryParams{
		RepositoryName: "",
		Force:          true,
	}
	err = cmInstance.DeleteRepository(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository name cannot be empty")

	// Try to delete with invalid repository name (backslash)
	params.RepositoryName = "invalid\\name"
	err = cmInstance.DeleteRepository(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "backslashes")

	// Try to delete with reserved name
	params.RepositoryName = "."
	err = cmInstance.DeleteRepository(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reserved")
}

func TestRepositoryDeleteNonexistent(t *testing.T) {
	// Setup test environment
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create CM instance
	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(setup.ConfigPath).
			WithConfig(config.NewManager(setup.ConfigPath)).
			WithLogger(logger.NewVerboseLogger()),
	})
	require.NoError(t, err)

	// Try to delete a nonexistent repository
	params := codemanager.DeleteRepositoryParams{
		RepositoryName: "nonexistent-repo",
		Force:          true,
	}
	err = cmInstance.DeleteRepository(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not found")
}

func TestRepositoryDeleteOutsideBaseDirectory(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "cm-e2e-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a repository outside the base directory
	outsideRepoPath := filepath.Join(tempDir, "outside-repo")
	err = os.MkdirAll(outsideRepoPath, 0755)
	require.NoError(t, err)

	// Initialize git repository
	err = exec.Command("git", "init", outsideRepoPath).Run()
	require.NoError(t, err)

	// Create a CM instance with a different base directory
	cmDir := filepath.Join(tempDir, ".cm")
	cfg := config.Config{
		RepositoriesDir: cmDir,
		WorkspacesDir:   filepath.Join(tempDir, "workspaces"),
		StatusFile:      filepath.Join(cmDir, "status.yaml"),
	}

	// Write the config file
	configPath := filepath.Join(tempDir, "config.yaml")
	configData, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(configPath),
	})
	require.NoError(t, err)

	// Manually add the repository to status (simulating a repository outside base dir)
	// We'll use the status manager directly to add a repository with a path outside the base directory
	statusManager := status.NewManager(fs.NewFS(), cfg)

	// Create initial status
	err = statusManager.CreateInitialStatus()
	require.NoError(t, err)

	// Add repository with path outside base directory
	repoName := "outside-repo"
	addParams := status.AddRepositoryParams{
		Path: outsideRepoPath,
		Remotes: map[string]status.Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
	}
	err = statusManager.AddRepository(repoName, addParams)
	require.NoError(t, err)

	// Verify the repository was added
	repositories, err := cmInstance.ListRepositories()
	require.NoError(t, err)
	require.Len(t, repositories, 1)
	require.Equal(t, repoName, repositories[0].Name)
	require.False(t, repositories[0].InRepositoriesDir) // Should be false since it's outside base dir

	// Delete the repository
	params := codemanager.DeleteRepositoryParams{
		RepositoryName: repoName,
		Force:          true,
	}
	err = cmInstance.DeleteRepository(params)
	require.NoError(t, err)

	// Verify the repository was removed from status
	repositories, err = cmInstance.ListRepositories()
	require.NoError(t, err)
	require.Len(t, repositories, 0)

	// Verify the directory still exists (should not be deleted since it's outside base dir)
	_, err = os.Stat(outsideRepoPath)
	require.NoError(t, err, "Repository directory should still exist since it's outside base directory")
}

func TestRepositoryDeleteWithEmptyParentCleanup(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "cm-e2e-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a CM instance with a specific base directory
	cmDir := filepath.Join(tempDir, ".cm")
	cfg := config.Config{
		RepositoriesDir: cmDir,
		WorkspacesDir:   filepath.Join(tempDir, "workspaces"),
		StatusFile:      filepath.Join(cmDir, "status.yaml"),
	}

	// Write the config file
	configPath := filepath.Join(tempDir, "config.yaml")
	configData, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, configData, 0644))

	cmInstance, err := codemanager.NewCodeManager(codemanager.NewCodeManagerParams{
		Dependencies: createE2EDependencies(configPath),
	})
	require.NoError(t, err)

	// Create a repository with nested directory structure
	// This will create: cmDir/github.com/user/repo/origin/main
	repoName := "github.com/user/repo"
	repoPath := filepath.Join(cmDir, "github.com", "user", "repo", "origin", "main")

	// Create the nested directory structure
	err = os.MkdirAll(repoPath, 0755)
	require.NoError(t, err)

	// Initialize git repository
	err = exec.Command("git", "init", repoPath).Run()
	require.NoError(t, err)

	// Add repository to status
	statusManager := status.NewManager(fs.NewFS(), cfg)
	err = statusManager.CreateInitialStatus()
	require.NoError(t, err)

	addParams := status.AddRepositoryParams{
		Path: repoPath,
		Remotes: map[string]status.Remote{
			"origin": {
				DefaultBranch: "main",
			},
		},
	}
	err = statusManager.AddRepository(repoName, addParams)
	require.NoError(t, err)

	// Verify the repository was added
	repositories, err := cmInstance.ListRepositories()
	require.NoError(t, err)
	require.Len(t, repositories, 1)
	require.Equal(t, repoName, repositories[0].Name)

	// Verify the nested directory structure exists
	require.DirExists(t, filepath.Join(cmDir, "github.com"))
	require.DirExists(t, filepath.Join(cmDir, "github.com", "user"))
	require.DirExists(t, filepath.Join(cmDir, "github.com", "user", "repo"))
	require.DirExists(t, filepath.Join(cmDir, "github.com", "user", "repo", "origin"))
	require.DirExists(t, filepath.Join(cmDir, "github.com", "user", "repo", "origin", "main"))

	// Delete the repository
	params := codemanager.DeleteRepositoryParams{
		RepositoryName: repoName,
		Force:          true,
	}
	err = cmInstance.DeleteRepository(params)
	require.NoError(t, err)

	// Verify the repository was removed from status
	repositories, err = cmInstance.ListRepositories()
	require.NoError(t, err)
	require.Len(t, repositories, 0)

	// Verify the repository directory was deleted
	require.NoDirExists(t, repoPath)

	// Verify that empty parent directories were cleaned up
	// The entire nested structure should be removed since it's now empty
	require.NoDirExists(t, filepath.Join(cmDir, "github.com", "user", "repo", "origin"))
	require.NoDirExists(t, filepath.Join(cmDir, "github.com", "user", "repo"))
	require.NoDirExists(t, filepath.Join(cmDir, "github.com", "user"))
	require.NoDirExists(t, filepath.Join(cmDir, "github.com"))

	// Verify that the base repositories directory still exists
	require.DirExists(t, cmDir)
}
