//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_ValidateRepositoryPath(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-validate-repo-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test empty path
	valid, err := fs.ValidateRepositoryPath("")
	assert.Error(t, err)
	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrInvalidRepository)

	// Test non-existent path
	nonExistentPath := filepath.Join(tmpDir, "non-existent")
	valid, err = fs.ValidateRepositoryPath(nonExistentPath)
	assert.Error(t, err)
	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrInvalidRepository)

	// Test path that exists but is a file (not directory)
	testFile := filepath.Join(tmpDir, "test-file.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(testFile)
	assert.Error(t, err)
	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrInvalidRepository)

	// Test directory without .git
	regularDir := filepath.Join(tmpDir, "regular-dir")
	err = os.MkdirAll(regularDir, 0755)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(regularDir)
	assert.Error(t, err)
	assert.False(t, valid)
	assert.ErrorIs(t, err, ErrInvalidRepository)

	// Test directory with .git as a file (submodule case)
	submoduleDir := filepath.Join(tmpDir, "submodule-dir")
	err = os.MkdirAll(submoduleDir, 0755)
	require.NoError(t, err)

	// Create .git as a file (submodule)
	gitFile := filepath.Join(submoduleDir, ".git")
	err = os.WriteFile(gitFile, []byte("gitdir: ../.git/modules/submodule"), 0644)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(submoduleDir)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test directory with .git as a directory (regular repository)
	repositoriesDir := filepath.Join(tmpDir, "repo-dir")
	err = os.MkdirAll(repositoriesDir, 0755)
	require.NoError(t, err)

	// Create .git directory
	gitDir := filepath.Join(repositoriesDir, ".git")
	err = os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	// Create some basic Git repository structure
	objectsDir := filepath.Join(gitDir, "objects")
	refsDir := filepath.Join(gitDir, "refs")
	err = os.MkdirAll(objectsDir, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(refsDir, 0755)
	require.NoError(t, err)

	// Create HEAD file
	headFile := filepath.Join(gitDir, "HEAD")
	err = os.WriteFile(headFile, []byte("ref: refs/heads/main\n"), 0644)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(repositoriesDir)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test nested repository
	nestedRepositoriesDir := filepath.Join(tmpDir, "nested", "repo")
	err = os.MkdirAll(nestedRepositoriesDir, 0755)
	require.NoError(t, err)

	// Create .git directory in nested location
	nestedGitDir := filepath.Join(nestedRepositoriesDir, ".git")
	err = os.MkdirAll(nestedGitDir, 0755)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(nestedRepositoriesDir)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test repository with absolute path
	absRepositoriesDir, err := filepath.Abs(repositoriesDir)
	require.NoError(t, err)

	valid, err = fs.ValidateRepositoryPath(absRepositoriesDir)
	assert.NoError(t, err)
	assert.True(t, valid)
}
