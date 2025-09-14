//go:build e2e

package test

import (
	"path/filepath"
	"testing"

	"github.com/lerenn/code-manager/pkg/cm"
	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryDelete(t *testing.T) {
	// Setup test environment
	setup := setupTestEnvironment(t)
	defer cleanupTestEnvironment(t, setup)

	// Create CM instance
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: config.Config{
			RepositoriesDir: setup.CmPath,
			StatusFile:      setup.StatusPath,
		},
		Logger: logger.NewVerboseLogger(),
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
	params := cm.DeleteRepositoryParams{
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
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: config.Config{
			RepositoriesDir: setup.CmPath,
			StatusFile:      setup.StatusPath,
			WorkspacesDir:   filepath.Join(setup.CmPath, "workspaces"),
		},
		Logger: logger.NewVerboseLogger(),
	})
	require.NoError(t, err)

	// Clone a test repository
	repoURL := "https://github.com/octocat/Hello-World.git"
	err = cmInstance.Clone(repoURL)
	require.NoError(t, err)

	// Create a workspace with the repository
	workspaceParams := cm.CreateWorkspaceParams{
		WorkspaceName: "test-workspace",
		Repositories:  []string{"github.com/octocat/Hello-World"},
	}
	err = cmInstance.CreateWorkspace(workspaceParams)
	require.NoError(t, err)

	// Try to delete the repository (should fail because it's in a workspace)
	deleteParams := cm.DeleteRepositoryParams{
		RepositoryName: "github.com/octocat/Hello-World",
		Force:          true,
	}
	err = cmInstance.DeleteRepository(deleteParams)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is part of workspace")

	// Delete the workspace first
	workspaceDeleteParams := cm.DeleteWorkspaceParams{
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
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: config.Config{
			RepositoriesDir: setup.CmPath,
			StatusFile:      setup.StatusPath,
		},
		Logger: logger.NewVerboseLogger(),
	})
	require.NoError(t, err)

	// Try to delete with empty repository name
	params := cm.DeleteRepositoryParams{
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
	cmInstance, err := cm.NewCM(cm.NewCMParams{
		Config: config.Config{
			RepositoriesDir: setup.CmPath,
			StatusFile:      setup.StatusPath,
		},
		Logger: logger.NewVerboseLogger(),
	})
	require.NoError(t, err)

	// Try to delete a nonexistent repository
	params := cm.DeleteRepositoryParams{
		RepositoryName: "nonexistent-repo",
		Force:          true,
	}
	err = cmInstance.DeleteRepository(params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository not found")
}
