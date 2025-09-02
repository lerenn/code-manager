//go:build unit

package workspace

import (
	"fmt"
	"testing"

	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/lerenn/code-manager/pkg/worktree"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWorkspace_DetectWorkspaceFiles_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompter(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
	})

	// Mock workspace file detection
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"test.code-workspace"}, nil)

	files, err := workspace.DetectWorkspaceFiles()
	assert.NoError(t, err)
	assert.Equal(t, []string{"test.code-workspace"}, files)
}

func TestWorkspace_DetectWorkspaceFiles_NoFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompter(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
	})

	// Mock workspace file detection - no files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	files, err := workspace.DetectWorkspaceFiles()
	assert.NoError(t, err)
	assert.Empty(t, files)
}

func TestWorkspace_DetectWorkspaceFiles_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompter(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
	})

	// Mock workspace file detection error
	mockFS.EXPECT().Glob("*.code-workspace").Return(nil, fmt.Errorf("glob error"))

	files, err := workspace.DetectWorkspaceFiles()
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.ErrorIs(t, err, ErrFailedToCheckWorkspaceFiles)
}

func TestWorkspace_ParseFile_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompter(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
	})

	// Mock workspace file content
	workspaceContent := `{
		"name": "Test Workspace",
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			},
			{
				"name": "Backend",
				"path": "./backend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("test.code-workspace").Return([]byte(workspaceContent), nil)

	config, err := workspace.ParseFile("test.code-workspace")
	assert.NoError(t, err)
	assert.Equal(t, "Test Workspace", config.Name)
	assert.Len(t, config.Folders, 2)
	assert.Equal(t, "Frontend", config.Folders[0].Name)
	assert.Equal(t, "./frontend", config.Folders[0].Path)
	assert.Equal(t, "Backend", config.Folders[1].Name)
	assert.Equal(t, "./backend", config.Folders[1].Path)
}

func TestWorkspace_ParseFile_EmptyFolders(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompter(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
	})

	// Mock workspace file with empty folders
	workspaceContent := `{
		"name": "Test Workspace",
		"folders": []
	}`
	mockFS.EXPECT().ReadFile("test.code-workspace").Return([]byte(workspaceContent), nil)

	config, err := workspace.ParseFile("test.code-workspace")
	assert.ErrorIs(t, err, ErrNoRepositoriesFound)
	assert.Nil(t, config)
}

func TestWorkspace_GetName_FromConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompter(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
	})

	config := &Config{
		Name: "Test Workspace",
		Folders: []Folder{
			{Path: "./frontend"},
		},
	}

	name := workspace.GetName(config, "test.code-workspace")
	assert.Equal(t, "Test Workspace", name)
}

func TestWorkspace_GetName_FromFilename(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompter(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
	})

	config := &Config{
		Folders: []Folder{
			{Path: "./frontend"},
		},
	}

	name := workspace.GetName(config, "my-project.code-workspace")
	assert.Equal(t, "my-project", name)
}
