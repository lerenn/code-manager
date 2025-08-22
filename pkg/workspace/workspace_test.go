//go:build unit

package workspace

import (
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

func TestNewWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})

	assert.NotNil(t, workspace)
}

func TestWorkspace_ListWorktrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})
	workspace.(*realWorkspace).OriginalFile = "/test/path/workspace.code-workspace"

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

	result, err := workspace.ListWorktrees(false)
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
	mockWorktree := worktree.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
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

	err := workspace.Load(false)
	assert.NoError(t, err)
	assert.Equal(t, "project.code-workspace", workspace.(*realWorkspace).OriginalFile)
}

func TestWorkspace_Load_NoFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})

	// Mock no workspace files found
	mockFS.EXPECT().Glob("*.code-workspace").Return([]string{}, nil)

	err := workspace.Load(false)
	assert.NoError(t, err)
	assert.Equal(t, "", workspace.(*realWorkspace).OriginalFile)
}

func TestWorkspace_Load_AlreadyLoaded(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fs.NewMockFS(ctrl)
	mockGit := git.NewMockGit(ctrl)
	mockStatus := status.NewMockManager(ctrl)
	mockLogger := logger.NewNoopLogger()
	mockPrompt := prompt.NewMockPrompt(ctrl)
	mockWorktree := worktree.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        mockLogger,
		Prompt:        mockPrompt,
		Worktree:      mockWorktree,
		Verbose:       true,
	})
	workspace.(*realWorkspace).OriginalFile = "already-loaded.code-workspace"

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

	err := workspace.Load(false)
	assert.NoError(t, err)
	assert.Equal(t, "already-loaded.code-workspace", workspace.(*realWorkspace).OriginalFile)
}
