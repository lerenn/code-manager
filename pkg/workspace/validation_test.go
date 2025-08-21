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

func TestWorkspace_ValidateWorkspaceReferences_Success(t *testing.T) {
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
	mockGit.EXPECT().GetRepositoryName("frontend").Return("github.com/lerenn/example", nil)

	// Mock repository not found in status (will be added)
	mockStatus.EXPECT().GetRepository("github.com/lerenn/example").Return(nil, status.ErrRepositoryNotFound)

	// Mock checking if origin remote exists
	mockGit.EXPECT().RemoteExists("frontend", "origin").Return(true, nil)

	// Mock getting remotes
	mockGit.EXPECT().GetRemoteURL("frontend", "origin").Return("https://github.com/lerenn/example.git", nil)

	// Mock getting default branch
	mockGit.EXPECT().GetDefaultBranch("https://github.com/lerenn/example.git").Return("main", nil)

	// Mock adding repository to status
	mockStatus.EXPECT().AddRepository("github.com/lerenn/example", gomock.Any()).Return(nil)

	// Mock getting repository after adding
	mockStatus.EXPECT().GetRepository("github.com/lerenn/example").Return(&status.Repository{
		Path: "frontend",
		Remotes: map[string]status.Remote{
			"origin": {DefaultBranch: "main"},
		},
		Worktrees: make(map[string]status.WorktreeInfo),
	}, nil)

	// Mock worktree operations
	mockWorktree.EXPECT().BuildPath("github.com/lerenn/example", "origin", "main").Return("/test/path/github.com/lerenn/example/origin/main")
	mockWorktree.EXPECT().Exists("frontend", "main").Return(false, nil)
	mockWorktree.EXPECT().Create(gomock.Any()).Return(nil)
	mockWorktree.EXPECT().AddToStatus(gomock.Any()).Return(nil)

	err := workspace.ValidateWorkspaceReferences()
	assert.NoError(t, err)
}

func TestWorkspace_ValidateWorkspaceReferences_RepositoryNotFound(t *testing.T) {
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
