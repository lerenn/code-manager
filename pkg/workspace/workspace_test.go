//go:build unit

package workspace

import (
	"fmt"
	"testing"

	"github.com/lerenn/code-manager/pkg/config"
	"github.com/lerenn/code-manager/pkg/fs"
	"github.com/lerenn/code-manager/pkg/git"
	"github.com/lerenn/code-manager/pkg/logger"
	"github.com/lerenn/code-manager/pkg/prompt"
	"github.com/lerenn/code-manager/pkg/status"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// createTestConfig creates a test configuration.
func createTestConfig() *config.Config {
	return &config.Config{
		BasePath:   "/test/base/path",
		StatusFile: "/test/status.yaml",
	}
}

func TestWorkspace_DetectWorkspaceFiles_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
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
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
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
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
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
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
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
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
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
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
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
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

	config := &Config{
		Folders: []Folder{
			{Path: "./frontend"},
		},
	}

	name := workspace.GetName(config, "my-project.code-workspace")
	assert.Equal(t, "my-project", name)
}

func TestWorkspace_Validate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})
	workspace.OriginalFile = "test.code-workspace"

	// Mock workspace file content
	workspaceContent := `{
		"folders": [
			{
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("test.code-workspace").Return([]byte(workspaceContent), nil)

	// Mock repository validation
	mockFS.EXPECT().Exists("frontend").Return(true, nil)
	mockFS.EXPECT().Exists("frontend/.git").Return(true, nil)
	mockGit.EXPECT().Status("frontend").Return("On branch main", nil)

	err := workspace.Validate()
	assert.NoError(t, err)
}

func TestWorkspace_Validate_RepositoryNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})
	workspace.OriginalFile = "test.code-workspace"

	// Mock workspace file content
	workspaceContent := `{
		"folders": [
			{
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("test.code-workspace").Return([]byte(workspaceContent), nil)

	// Mock repository not found
	mockFS.EXPECT().Exists("frontend").Return(false, nil)

	err := workspace.Validate()
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}

func TestWorkspace_Validate_NoGitDirectory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})
	workspace.OriginalFile = "test.code-workspace"

	// Mock workspace file content
	workspaceContent := `{
		"folders": [
			{
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("test.code-workspace").Return([]byte(workspaceContent), nil)

	// Mock repository exists but no .git directory
	mockFS.EXPECT().Exists("frontend").Return(true, nil)
	mockFS.EXPECT().Exists("frontend/.git").Return(false, nil)

	err := workspace.Validate()
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}

func TestWorkspace_ListWorktrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})
	workspace.OriginalFile = "/test/path/workspace.code-workspace"

	// Mock GetWorkspace call
	workspaceInfo := &status.Workspace{
		Repositories: []string{"github.com/example/repo1", "github.com/example/repo2"},
	}
	mockStatus.EXPECT().GetWorkspace("/test/path/workspace.code-workspace").Return(workspaceInfo, nil)

	// Mock GetRepository calls - each repository can be called multiple times
	repo1 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:feature/test-branch": {
				Remote: "origin",
				Branch: "feature/test-branch",
			},
		},
	}
	repo2 := &status.Repository{
		Worktrees: map[string]status.WorktreeInfo{
			"origin:bugfix/issue-123": {
				Remote: "origin",
				Branch: "bugfix/issue-123",
			},
		},
	}
	// Each repository can be called multiple times as the algorithm checks all repos for each worktree
	mockStatus.EXPECT().GetRepository("github.com/example/repo1").Return(repo1, nil).AnyTimes()
	mockStatus.EXPECT().GetRepository("github.com/example/repo2").Return(repo2, nil).AnyTimes()

	result, err := workspace.ListWorktrees()
	assert.NoError(t, err)
	assert.Len(t, result, 2, "Should only return worktrees for current workspace")

	// Verify only current workspace worktrees are returned
	// Note: WorktreeInfo doesn't have Workspace field, so we can't verify this directly
	// The filtering is done internally by the workspace implementation

	// Verify specific branches are present
	branchNames := make([]string, len(result))
	for i, wt := range result {
		branchNames[i] = wt.Branch
	}
	assert.Contains(t, branchNames, "feature/test-branch")
	assert.Contains(t, branchNames, "bugfix/issue-123")
	assert.NotContains(t, branchNames, "feature/other-branch")
}

func TestWorkspace_Load_SingleFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

	// Mock single workspace file found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{"project.code-workspace"}, nil)

	// Mock workspace file content
	workspaceContent := `{
		"name": "Test Project",
		"folders": [
			{
				"name": "Frontend",
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("project.code-workspace").Return([]byte(workspaceContent), nil)

	err := workspace.Load()
	assert.NoError(t, err)
	assert.Equal(t, "project.code-workspace", workspace.OriginalFile)
}

func TestWorkspace_Load_NoFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})

	// Mock no workspace files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	err := workspace.Load()
	assert.NoError(t, err)
	assert.Equal(t, "", workspace.OriginalFile)
}

func TestWorkspace_Load_AlreadyLoaded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Verbose:       true,
	})
	workspace.OriginalFile = "already-loaded.code-workspace"

	// Mock workspace file content (should be called even if already loaded)
	workspaceContent := `{
		"name": "Already Loaded",
		"folders": [
			{
				"path": "./frontend"
			}
		]
	}`
	mockFS.EXPECT().ReadFile("already-loaded.code-workspace").Return([]byte(workspaceContent), nil)

	err := workspace.Load()
	assert.NoError(t, err)
	assert.Equal(t, "already-loaded.code-workspace", workspace.OriginalFile)
}
