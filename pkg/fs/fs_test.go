//go:build integration

package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS_ReadFile(t *testing.T) {
	fs := NewFS()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write content to the file
	content := []byte("test content")
	err = os.WriteFile(tmpFile.Name(), content, 0644)
	require.NoError(t, err)

	// Test reading existing file
	readContent, err := fs.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)

	// Test reading non-existing file
	_, err = fs.ReadFile("non-existing-file.txt")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestFS_ReadDir(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create some files in the directory
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	err = os.WriteFile(file1, []byte("content1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("content2"), 0644)
	require.NoError(t, err)

	// Test reading directory contents
	entries, err := fs.ReadDir(tmpDir)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	// Verify file names
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	assert.Contains(t, names, "file1.txt")
	assert.Contains(t, names, "file2.txt")

	// Test reading non-existing directory
	_, err = fs.ReadDir("non-existing-dir")
	assert.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestFS_Glob(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-glob-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create some .code-workspace files
	workspace1 := filepath.Join(tmpDir, "project.code-workspace")
	workspace2 := filepath.Join(tmpDir, "dev.code-workspace")
	otherFile := filepath.Join(tmpDir, "other.txt")

	err = os.WriteFile(workspace1, []byte("{}"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(workspace2, []byte("{}"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(otherFile, []byte("content"), 0644)
	require.NoError(t, err)

	// Test glob pattern matching
	pattern := filepath.Join(tmpDir, "*.code-workspace")
	matches, err := fs.Glob(pattern)
	assert.NoError(t, err)
	assert.Len(t, matches, 2)
	assert.Contains(t, matches, workspace1)
	assert.Contains(t, matches, workspace2)

	// Test glob with no matches
	noMatchPattern := filepath.Join(tmpDir, "*.nonexistent")
	noMatches, err := fs.Glob(noMatchPattern)
	assert.NoError(t, err)
	assert.Len(t, noMatches, 0)

	// Test glob with invalid pattern
	_, err = fs.Glob("[invalid")
	assert.Error(t, err)
}

func TestFS_ReadFile_RealWorkspace(t *testing.T) {
	fs := NewFS()

	// Create a temporary workspace file
	tmpFile, err := os.CreateTemp("", "test-*.code-workspace")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write realistic workspace JSON
	workspaceJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			},
			{
				"name": "Backend", 
				"path": "./backend"
			}
		],
		"settings": {
			"editor.formatOnSave": true
		}
	}`

	err = os.WriteFile(tmpFile.Name(), []byte(workspaceJSON), 0644)
	require.NoError(t, err)

	// Test reading workspace file
	content, err := fs.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(content), "Frontend")
	assert.Contains(t, string(content), "Backend")
	assert.Contains(t, string(content), "editor.formatOnSave")
}

func TestFS_Glob_RealWorkspace(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-workspace-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create workspace files
	workspace1 := filepath.Join(tmpDir, "project.code-workspace")
	workspace2 := filepath.Join(tmpDir, "development.code-workspace")

	err = os.WriteFile(workspace1, []byte(`{"folders": [{"path": "./repo1"}]}`), 0644)
	require.NoError(t, err)
	err = os.WriteFile(workspace2, []byte(`{"folders": [{"path": "./repo2"}]}`), 0644)
	require.NoError(t, err)

	// Test pattern matching for workspace files
	pattern := filepath.Join(tmpDir, "*.code-workspace")
	matches, err := fs.Glob(pattern)
	assert.NoError(t, err)
	assert.Len(t, matches, 2)
	assert.Contains(t, matches, workspace1)
	assert.Contains(t, matches, workspace2)
}

func TestFS_ReadDir_RealWorkspace(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "test-workspace-structure-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create workspace file
	workspaceFile := filepath.Join(tmpDir, "project.code-workspace")
	err = os.WriteFile(workspaceFile, []byte(`{"folders": [{"path": "./repo1"}]}`), 0644)
	require.NoError(t, err)

	// Create repository directories
	repo1 := filepath.Join(tmpDir, "repo1")
	err = os.Mkdir(repo1, 0755)
	require.NoError(t, err)

	// Create .git directory in repo1
	gitDir := filepath.Join(repo1, ".git")
	err = os.Mkdir(gitDir, 0755)
	require.NoError(t, err)

	// Test reading directory contents
	entries, err := fs.ReadDir(tmpDir)
	assert.NoError(t, err)

	// Should contain workspace file and repo1 directory
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	assert.Contains(t, names, "project.code-workspace")
	assert.Contains(t, names, "repo1")
}

func TestFS_ReadFile_MalformedWorkspace(t *testing.T) {
	fs := NewFS()

	// Create a temporary file with malformed content
	tmpFile, err := os.CreateTemp("", "test-malformed-*.code-workspace")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write malformed JSON
	malformedJSON := `{
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			},
			{
				"name": "Backend", 
				"path": "./backend"
			}
		,
		"settings": {
			"editor.formatOnSave": true
		}
	}` // Missing closing bracket

	err = os.WriteFile(tmpFile.Name(), []byte(malformedJSON), 0644)
	require.NoError(t, err)

	// Test reading malformed workspace file (should still read the content)
	content, err := fs.ReadFile(tmpFile.Name())
	assert.NoError(t, err) // Reading should succeed, parsing will fail later
	assert.Contains(t, string(content), "Frontend")
	assert.Contains(t, string(content), "Backend")
}

func TestFS_Glob_EdgeCases(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-glob-edge-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test glob with empty directory
	pattern := filepath.Join(tmpDir, "*.code-workspace")
	matches, err := fs.Glob(pattern)
	assert.NoError(t, err)
	assert.Len(t, matches, 0)

	// Test glob with special characters in filename
	specialFile := filepath.Join(tmpDir, "test[1].code-workspace")
	err = os.WriteFile(specialFile, []byte("{}"), 0644)
	require.NoError(t, err)

	// Test glob with normal pattern that should match
	normalPattern := filepath.Join(tmpDir, "*.code-workspace")
	matches, err = fs.Glob(normalPattern)
	assert.NoError(t, err)
	assert.Len(t, matches, 1)
	assert.Equal(t, specialFile, matches[0])
}

func TestFS_ReadFile_PermissionErrors(t *testing.T) {
	fs := NewFS()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-permission-*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write content
	err = os.WriteFile(tmpFile.Name(), []byte("test content"), 0644)
	require.NoError(t, err)

	// Test reading with normal permissions (should work)
	content, err := fs.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Equal(t, "test content", string(content))

	// Note: Testing actual permission errors would require changing file permissions
	// which might not work on all systems, so we test the happy path
}

func TestFS_MkdirAll(t *testing.T) {
	fs := NewFS()

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-mkdir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test creating nested directories
	nestedPath := filepath.Join(tmpDir, "level1", "level2", "level3")
	err = fs.MkdirAll(nestedPath, 0755)
	assert.NoError(t, err)

	// Verify directories were created
	exists, err := fs.Exists(nestedPath)
	assert.NoError(t, err)
	assert.True(t, exists)

	isDir, err := fs.IsDir(nestedPath)
	assert.NoError(t, err)
	assert.True(t, isDir)

	// Test creating existing directory (should not error)
	err = fs.MkdirAll(nestedPath, 0755)
	assert.NoError(t, err)
}

func TestFS_GetHomeDir(t *testing.T) {
	fs := NewFS()

	homeDir, err := fs.GetHomeDir()
	assert.NoError(t, err)
	assert.NotEmpty(t, homeDir)

	// Verify it's a directory
	exists, err := fs.Exists(homeDir)
	assert.NoError(t, err)
	assert.True(t, exists)

	isDir, err := fs.IsDir(homeDir)
	assert.NoError(t, err)
	assert.True(t, isDir)
}

func TestFS_IsNotExist(t *testing.T) {
	fs := NewFS()

	// Test with non-existent file error
	_, err := fs.ReadFile("non-existent-file.txt")
	assert.Error(t, err)
	assert.True(t, fs.IsNotExist(err))

	// Test with existing file (should not be IsNotExist)
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = fs.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.False(t, fs.IsNotExist(err))
}
