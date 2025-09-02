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

func TestWorkspace_ValidateWorkspaceReferences_Success(t *testing.T) {
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
	workspace.(*realWorkspace).OriginalFile = "test.code-workspace"

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

	// Mock repository URL extraction
	mockGit.EXPECT().GetRepositoryName("frontend").Return("github.com/octocat/Hello-World", nil)

	// Mock repository not found in status (will be added)
	mockStatus.EXPECT().GetRepository("github.com/octocat/Hello-World").Return(nil, status.ErrRepositoryNotFound)

	// Mock checking if origin remote exists
	mockGit.EXPECT().RemoteExists("frontend", "origin").Return(true, nil)

	// Mock getting remotes
	mockGit.EXPECT().GetRemoteURL("frontend", "origin").Return("https://github.com/octocat/Hello-World.git", nil)

	// Mock getting default branch
	mockGit.EXPECT().GetDefaultBranch("https://github.com/octocat/Hello-World.git").Return("main", nil)

	// Mock adding repository to status
	mockStatus.EXPECT().AddRepository("github.com/octocat/Hello-World", gomock.Any()).Return(nil)

	// Mock getting repository after adding
	mockStatus.EXPECT().GetRepository("github.com/octocat/Hello-World").Return(&status.Repository{
		Path: "frontend",
		Remotes: map[string]status.Remote{
			"origin": {DefaultBranch: "main"},
		},
		Worktrees: make(map[string]status.WorktreeInfo),
	}, nil)

	// Mock worktree operations
	mockWorktree.EXPECT().BuildPath("github.com/octocat/Hello-World", "origin", "main").Return("/test/path/github.com/octocat/Hello-World/origin/main")
	mockWorktree.EXPECT().Exists("frontend", "main").Return(false, nil)
	mockWorktree.EXPECT().Create(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	err := workspace.ValidateWorkspaceReferences()
	assert.NoError(t, err)
}

func TestWorkspace_ValidateWorkspaceReferences_RepositoryNotFound(t *testing.T) {
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
	workspace.(*realWorkspace).OriginalFile = "test.code-workspace"

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

	err := workspace.ValidateWorkspaceReferences()
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}

func TestWorkspace_ValidateWorkspaceReferences_NoGitDirectory(t *testing.T) {
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
	workspace.(*realWorkspace).OriginalFile = "test.code-workspace"

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

	err := workspace.ValidateWorkspaceReferences()
	assert.ErrorIs(t, err, ErrRepositoryNotFound)
}
