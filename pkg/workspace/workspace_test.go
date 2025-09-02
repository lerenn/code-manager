//go:build unit

package workspace

import (
	"testing"

	fsmocks "github.com/lerenn/code-manager/pkg/fs/mocks"
	gitmocks "github.com/lerenn/code-manager/pkg/git/mocks"
	"github.com/lerenn/code-manager/pkg/logger"
	promptmocks "github.com/lerenn/code-manager/pkg/prompt/mocks"
	"github.com/lerenn/code-manager/pkg/status"
	statusmocks "github.com/lerenn/code-manager/pkg/status/mocks"
	"github.com/lerenn/code-manager/pkg/worktree"
	worktreemocks "github.com/lerenn/code-manager/pkg/worktree/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewWorkspace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        logger.NewNoopLogger(),
		Prompt:        mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree {
			return mockWorktree
		},
	})

	assert.NotNil(t, workspace)
}

func TestWorkspace_ListWorktrees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        logger.NewNoopLogger(),
		Prompt:        mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree {
			return mockWorktree
		},
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

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        logger.NewNoopLogger(),
		Prompt:        mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree {
			return mockWorktree
		},
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

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        logger.NewNoopLogger(),
		Prompt:        mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree {
			return mockWorktree
		},
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

	mockFS := fsmocks.NewMockFS(ctrl)
	mockGit := gitmocks.NewMockGit(ctrl)
	mockStatus := statusmocks.NewMockManager(ctrl)

	mockPrompt := promptmocks.NewMockPrompter(ctrl)
	mockWorktree := worktreemocks.NewMockWorktree(ctrl)

	workspace := NewWorkspace(NewWorkspaceParams{
		FS:            mockFS,
		Git:           mockGit,
		Config:        createTestConfig(),
		StatusManager: mockStatus,
		Logger:        logger.NewNoopLogger(),
		Prompt:        mockPrompt,
		WorktreeProvider: func(params worktree.NewWorktreeParams) worktree.Worktree {
			return mockWorktree
		},
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
